package handlers

import (
	"net/http"

	"memoryconnect/middleware"
	"memoryconnect/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}

type ledgerDTO struct {
	ID           uint   `json:"id"`
	Change       int    `json:"change"`
	BalanceAfter int    `json:"balanceAfter"`
	FrozenAfter  int    `json:"frozenAfter"`
	Type         string `json:"type"`
	TypeText     string `json:"typeText"`
	StoryID      *uint  `json:"storyId"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"createdAt"`
}

var ledgerTypeText = map[string]string{
	"register": "注册赠送",
	"freeze":   "发布冻结",
	"append":   "追加冻结",
	"pay":      "结算支出",
	"income":   "代看收入",
}

type briefStory struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	City      string `json:"city"`
	Reward    int    `json:"reward"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// Me 当前登录用户的完整信息 + 统计
func (h *UserHandler) Me(c *gin.Context) {
	u := middleware.CurrentUser(c)

	var published int64
	h.DB.Model(&models.Story{}).Where("author_id = ?", u.ID).Count(&published)
	var settled int64
	h.DB.Model(&models.Story{}).Where("author_id = ? AND status = ?", u.ID, "settled").Count(&settled)
	var helped int64
	h.DB.Model(&models.Response{}).Where("user_id = ? AND accepted = ?", u.ID, true).Count(&helped)
	var earned int64
	h.DB.Model(&models.CoinLedger{}).Where("user_id = ? AND type = ?", u.ID, "income").Count(&earned)

	c.JSON(http.StatusOK, gin.H{
		"user":          userJSON(u),
		"published":     published,
		"settled":       settled,
		"helped":        helped,
		"incomeCount":   earned,
	})
}

// Ledger 个人硬币流水
func (h *UserHandler) Ledger(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var rows []models.CoinLedger
	h.DB.Where("user_id = ?", u.ID).Order("created_at DESC").Limit(200).Find(&rows)

	out := make([]ledgerDTO, 0, len(rows))
	for _, l := range rows {
		sid := l.StoryID
		out = append(out, ledgerDTO{
			ID:           l.ID,
			Change:       l.Change,
			BalanceAfter: l.BalanceAfter,
			FrozenAfter:  l.FrozenAfter,
			Type:         l.Type,
			TypeText:     ledgerTypeText[l.Type],
			StoryID:      sid,
			Remark:       l.Remark,
			CreatedAt:    l.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	c.JSON(http.StatusOK, gin.H{"ledgers": out})
}

// MyStories 我发布的求看（带是否已被回应/采纳的简要信息）
func (h *UserHandler) MyStories(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var stories []models.Story
	h.DB.Where("author_id = ?", u.ID).Order("created_at DESC").Find(&stories)

	out := make([]gin.H, 0, len(stories))
	for _, s := range stories {
		var respCount int64
		h.DB.Model(&models.Response{}).Where("story_id = ?", s.ID).Count(&respCount)
		out = append(out, gin.H{
			"id":           s.ID,
			"title":        s.Title,
			"city":         s.City,
			"reward":       s.Reward,
			"status":       s.Status,
			"responseCount": respCount,
			"createdAt":    s.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	c.JSON(http.StatusOK, gin.H{"stories": out})
}

// MyResponses 我替别人拍过的现场
func (h *UserHandler) MyResponses(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var responses []models.Response
	h.DB.Where("user_id = ?", u.ID).Order("created_at DESC").Find(&responses)

	storyIDs := []uint{}
	for _, r := range responses {
		storyIDs = append(storyIDs, r.StoryID)
	}
	storyMap := map[uint]models.Story{}
	if len(storyIDs) > 0 {
		var stories []models.Story
		h.DB.Where("id IN ?", storyIDs).Find(&stories)
		for _, s := range stories {
			storyMap[s.ID] = s
		}
	}

	out := make([]gin.H, 0, len(responses))
	for _, r := range responses {
		s := storyMap[r.StoryID]
		var author models.User
		h.DB.First(&author, s.AuthorID)
		out = append(out, gin.H{
			"responseId": r.ID,
			"storyId":    r.StoryID,
			"title":      s.Title,
			"city":       s.City,
			"reward":     s.Reward,
			"accepted":   r.Accepted,
			"status":     s.Status,
			"author":     brief(author),
			"createdAt":  r.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	c.JSON(http.StatusOK, gin.H{"responses": out})
}
