package handlers

import (
	"net/http"

	"memorylink/middleware"
	"memorylink/models"

	"github.com/gin-gonic/gin"
)

// Profile 个人中心：余额 + 我发布的 + 我代看的 + 硬币流水
func Profile(c *gin.Context) {
	uid := middleware.CurrentUserID(c)

	var user models.User
	if err := models.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	var myStories []models.Story
	models.DB.Preload("User").Preload("Responses.User").
		Where("user_id = ?", uid).Order("created_at DESC").Find(&myStories)

	var myResponses []models.Response
	models.DB.Preload("User").Preload("Story.User").
		Where("user_id = ?", uid).Order("created_at DESC").Find(&myResponses)

	var ledgers []models.CoinLedger
	models.DB.Where("user_id = ?", uid).Order("created_at DESC").Limit(100).Find(&ledgers)

	// 统计冻结中的硬币（未完成故事的悬赏总额）
	var frozen int64
	models.DB.Model(&models.Story{}).
		Where("user_id = ? AND status = ?", uid, models.StoryStatusOpen).
		Select("COALESCE(SUM(bounty),0)").Scan(&frozen)

	var earned int64
	models.DB.Model(&models.CoinLedger{}).
		Where("user_id = ? AND type = ?", uid, models.CoinTypeReward).
		Select("COALESCE(SUM(`change`),0)").Scan(&earned)

	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"frozen":        frozen,
		"earned":        earned,
		"my_stories":    myStories,
		"my_responses":  myResponses,
		"coin_ledgers":  ledgers,
	})
}

// UserStories 查看某用户发布的求看（公开主页用）
func UserStories(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := models.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	var stories []models.Story
	models.DB.Preload("Responses.User").
		Where("user_id = ?", id).Order("created_at DESC").Find(&stories)
	c.JSON(http.StatusOK, gin.H{"user": user, "stories": stories})
}
