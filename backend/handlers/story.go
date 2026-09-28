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

type storyReq struct {
	Title      string   `json:"title" binding:"required,max=100"`
	City       string   `json:"city" binding:"required,max=50"`
	Location   string   `json:"location" binding:"required,max=200"`
	MemoryText string   `json:"memory_text" binding:"required"`
	OldPhotos  []string `json:"old_photos"`
	Bounty     int      `json:"bounty" binding:"required,min=1"`
}

// CreateStory 发布求看：事务内冻结（扣除）悬赏硬币
func CreateStory(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var req storyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整的标题、城市、地点、回忆和悬赏硬币"})
		return
	}
	if req.Bounty <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "悬赏至少 1 枚硬币"})
		return
	}

	story := models.Story{
		UserID:     uid,
		Title:      strings.TrimSpace(req.Title),
		City:       strings.TrimSpace(req.City),
		Location:   strings.TrimSpace(req.Location),
		MemoryText: strings.TrimSpace(req.MemoryText),
		OldPhotos:  strings.Join(trimURLs(req.OldPhotos), ","),
		Bounty:     req.Bounty,
		Status:     models.StoryStatusOpen,
	}

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		// 行锁锁定用户，校验余额
		var user models.User
		if err := tx.Clauses(lockedForUpdate()).First(&user, uid).Error; err != nil {
			return err
		}
		if user.Balance < req.Bounty {
			return errInsufficientCoins
		}
		// 扣除冻结
		if err := tx.Model(&user).Update("balance", gorm.Expr("balance - ?", req.Bounty)).Error; err != nil {
			return err
		}
		if err := tx.Create(&story).Error; err != nil {
			return err
		}
		return writeLedger(tx, uid, -req.Bounty, models.CoinTypeFreeze,
			"发布求看《"+story.Title+"》冻结悬赏", &story.ID, nil)
	})
	if err != nil {
		if errors.Is(err, errInsufficientCoins) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "记忆硬币余额不足，无法冻结悬赏"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
		return
	}

	loadStoryAssociations(&story)
	c.JSON(http.StatusOK, gin.H{"story": story})
}

// ListStories Feed 流：求看需求列表
func ListStories(c *gin.Context) {
	var stories []models.Story
	q := models.DB.Preload("User").Preload("Responses.User").
		Order("created_at DESC")

	if city := c.Query("city"); city != "" {
		q = q.Where("city = ?", city)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if mine := c.Query("mine"); mine == "1" {
		uid := middleware.CurrentUserID(c)
		q = q.Where("user_id = ?", uid)
	}
	if err := q.Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stories": stories})
}

// GetStory 故事详情
func GetStory(c *gin.Context) {
	id := c.Param("id")
	var story models.Story
	if err := models.DB.Preload("User").Preload("Responses.User").
		First(&story, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "故事不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"story": story})
}

type appendReq struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

// AppendBounty 追加悬赏：事务内继续冻结硬币
func AppendBounty(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	id := c.Param("id")
	var req appendReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入正确的追加数量"})
		return
	}

	var story models.Story
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(lockedForUpdate()).First(&story, id).Error; err != nil {
			return err
		}
		if story.UserID != uid {
			return errNotOwner
		}
		if story.Status != models.StoryStatusOpen {
			return errStoryClosed
		}
		var user models.User
		if err := tx.Clauses(lockedForUpdate()).First(&user, uid).Error; err != nil {
			return err
		}
		if user.Balance < req.Amount {
			return errInsufficientCoins
		}
		if err := tx.Model(&user).Update("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&story).Update("bounty", gorm.Expr("bounty + ?", req.Amount)).Error; err != nil {
			return err
		}
		return writeLedger(tx, uid, -req.Amount, models.CoinTypeAdd,
			"追加求看《"+story.Title+"》悬赏冻结", &story.ID, nil)
	})
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "故事不存在"})
		case errors.Is(err, errNotOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "只能追加自己发布的求看悬赏"})
		case errors.Is(err, errStoryClosed):
			c.JSON(http.StatusConflict, gin.H{"error": "该求看已结束，无法继续追加"})
		case errors.Is(err, errInsufficientCoins):
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "记忆硬币余额不足"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "追加失败"})
		}
		return
	}

	models.DB.Preload("User").Preload("Responses.User").First(&story, story.ID)
	c.JSON(http.StatusOK, gin.H{"story": story})
}

func trimURLs(urls []string) []string {
	out := make([]string, 0, len(urls))
	for _, u := range urls {
		if u = strings.TrimSpace(u); u != "" {
			out = append(out, u)
		}
	}
	return out
}

func loadStoryAssociations(story *models.Story) {
	models.DB.Preload("User").Preload("Responses.User").First(story, story.ID)
}
