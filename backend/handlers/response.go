package handlers

import (
	"errors"
	"net/http"
	"strings"

	"memorylink/middleware"
	"memorylink/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type responseReq struct {
	NowText   string   `json:"now_text" binding:"required"`
	NewPhotos []string `json:"new_photos"`
	Message   string   `json:"message"`
}

// CreateResponse 替他去拍：上传现场新照与寄语
func CreateResponse(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	storyID := c.Param("id")

	var req responseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请写下你在现场看到的样子"})
		return
	}

	var story models.Story
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		// 锁定故事行，防止提交过程中状态变化
		if err := tx.Clauses(lockedForUpdate()).First(&story, storyID).Error; err != nil {
			return err
		}
		if story.Status != models.StoryStatusOpen {
			return errStoryClosed
		}
		if story.UserID == uid {
			return errors.New("不能替自己代看")
		}
		// 同一用户对同一故事只能提交一次
		var n int64
		tx.Model(&models.Response{}).
			Where("story_id = ? AND user_id = ?", story.ID, uid).Count(&n)
		if n > 0 {
			return errors.New("你已提交过现场记录")
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "故事不存在"})
		case errors.Is(err, errStoryClosed):
			c.JSON(http.StatusConflict, gin.H{"error": "这个求看已经被采纳完成了"})
		default:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		}
		return
	}

	resp := models.Response{
		StoryID:   story.ID,
		UserID:    uid,
		NowText:   strings.TrimSpace(req.NowText),
		NewPhotos: strings.Join(trimURLs(req.NewPhotos), ","),
		Message:   strings.TrimSpace(req.Message),
		Status:    models.RespStatusPending,
	}
	if err := models.DB.Create(&resp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交失败"})
		return
	}
	models.DB.Preload("User").First(&resp, resp.ID)
	c.JSON(http.StatusOK, gin.H{"response": resp})
}

// AcceptResponse 发起人确认采纳：事务内结算冻结的悬赏硬币给代看人
func AcceptResponse(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	respID := c.Param("rid")

	var resp models.Response
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 锁定回应与故事，保证并发下只能被采纳一次
		if err := tx.Clauses(lockedForUpdate()).First(&resp, respID).Error; err != nil {
			return err
		}
		var story models.Story
		if err := tx.Clauses(lockedForUpdate()).First(&story, resp.StoryID).Error; err != nil {
			return err
		}
		if story.UserID != uid {
			return errNotOwner
		}
		if story.Status != models.StoryStatusOpen {
			return errStoryClosed
		}
		if resp.Status != models.RespStatusPending {
			return errNotPending
		}

		bounty := story.Bounty
		if bounty <= 0 {
			return errors.New("悬赏金额异常")
		}

		// 2. 硬币结算：悬赏在发布时已从发起人余额冻结，
		//    采纳时直接发放给代看人（无需再扣发起人）。
		var responder models.User
		if err := tx.Clauses(lockedForUpdate()).First(&responder, resp.UserID).Error; err != nil {
			return err
		}
		if err := tx.Model(&responder).
			Update("balance", gorm.Expr("balance + ?", bounty)).Error; err != nil {
			return err
		}

		// 3. 更新状态：回应 accepted、故事 fulfilled
		if err := tx.Model(&resp).Updates(map[string]interface{}{
			"status": models.RespStatusAccepted,
		}).Error; err != nil {
			return err
		}
		// 其它待确认回应标记为 rejected
		if err := tx.Model(&models.Response{}).
			Where("story_id = ? AND id <> ? AND status = ?", story.ID, resp.ID, models.RespStatusPending).
			Update("status", models.RespStatusRejected).Error; err != nil {
			return err
		}
		if err := tx.Model(&story).Updates(map[string]interface{}{
			"status":               models.StoryStatusFulfilled,
			"accepted_response_id": resp.ID,
		}).Error; err != nil {
			return err
		}

		// 4. 代看人收入流水
		if err := writeLedger(tx, resp.UserID, bounty, models.CoinTypeReward,
			"代看《"+story.Title+"》被采纳，获得悬赏", &story.ID, &resp.ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "回应不存在"})
		case errors.Is(err, errNotOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "只有发起人能确认采纳"})
		case errors.Is(err, errStoryClosed):
			c.JSON(http.StatusConflict, gin.H{"error": "该求看已完成结算"})
		case errors.Is(err, errNotPending):
			c.JSON(http.StatusConflict, gin.H{"error": "该回应已处理过"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "结算失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已确认采纳，记忆硬币已结算给对方", "response_id": resp.ID})
}
