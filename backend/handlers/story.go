package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"memoryconnect/config"
	"memoryconnect/middleware"
	"memoryconnect/models"
	"memoryconnect/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StoryHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

// ---------- 请求体 ----------

type createStoryRequest struct {
	Title     string   `json:"title" binding:"required,max=100"`
	City      string   `json:"city"`
	Location  string   `json:"location"`
	Memory    string   `json:"memory"`
	OldPhotos []string `json:"oldPhotos"`
	Wish      string   `json:"wish"`
	Reward    int      `json:"reward" binding:"required,min=1"`
}

type appendRewardRequest struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

type createResponseRequest struct {
	Message   string   `json:"message" binding:"required,max=2000"`
	NewPhotos []string `json:"newPhotos"`
}

// ---------- 响应 DTO ----------

type userBrief struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type responseDTO struct {
	ID        uint      `json:"id"`
	StoryID   uint      `json:"storyId"`
	User      userBrief `json:"user"`
	Message   string    `json:"message"`
	NewPhotos []string  `json:"newPhotos"`
	Accepted  bool      `json:"accepted"`
	CreatedAt string    `json:"createdAt"`
}

type storyDTO struct {
	ID         uint          `json:"id"`
	Title      string        `json:"title"`
	City       string        `json:"city"`
	Location   string        `json:"location"`
	Memory     string        `json:"memory"`
	OldPhotos  []string      `json:"oldPhotos"`
	Wish       string        `json:"wish"`
	Reward     int           `json:"reward"`
	Status     string        `json:"status"`
	Author     userBrief     `json:"author"`
	Responses  []responseDTO `json:"responses"`
	HasMine    bool          `json:"hasMine"`
	MyResponse *uint         `json:"myResponseId"`
	CanSettle  bool          `json:"canSettle"`
	CreatedAt  string        `json:"createdAt"`
}

func splitPhotos(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func brief(u models.User) userBrief {
	nick := u.Nickname
	if nick == "" {
		nick = u.Username
	}
	return userBrief{ID: u.ID, Nickname: nick, Avatar: u.Avatar}
}

func fetchUsers(db *gorm.DB, ids []uint) map[uint]models.User {
	m := map[uint]models.User{}
	uniq := map[uint]bool{}
	ids2 := []uint{}
	for _, id := range ids {
		if id != 0 && !uniq[id] {
			uniq[id] = true
			ids2 = append(ids2, id)
		}
	}
	if len(ids2) == 0 {
		return m
	}
	var users []models.User
	db.Where("id IN ?", ids2).Find(&users)
	for _, u := range users {
		m[u.ID] = u
	}
	return m
}

// buildStoryDTO 组装卡片，viewer 用于判断当前用户能否结算/是否已回应
func (h *StoryHandler) buildStoryDTO(st models.Story, responses []models.Response, viewer *models.User) storyDTO {
	userIDs := []uint{st.AuthorID}
	for _, r := range responses {
		userIDs = append(userIDs, r.UserID)
	}
	users := fetchUsers(h.DB, userIDs)

	dtos := make([]responseDTO, 0, len(responses))
	var myResponseID *uint
	hasMine := false
	for _, r := range responses {
		rid := r.ID
		d := responseDTO{
			ID:        r.ID,
			StoryID:   r.StoryID,
			User:      brief(users[r.UserID]),
			Message:   r.Message,
			NewPhotos: splitPhotos(r.NewPhotos),
			Accepted:  r.Accepted,
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04"),
		}
		dtos = append(dtos, d)
		if viewer != nil && r.UserID == viewer.ID {
			hasMine = true
			myResponseID = &rid
		}
	}

	canSettle := false
	if viewer != nil && st.AuthorID == viewer.ID && st.Status == "open" {
		for _, r := range responses {
			if r.UserID != viewer.ID {
				canSettle = true
				break
			}
		}
	}

	return storyDTO{
		ID:         st.ID,
		Title:      st.Title,
		City:       st.City,
		Location:   st.Location,
		Memory:     st.Memory,
		OldPhotos:  splitPhotos(st.OldPhotos),
		Wish:       st.Wish,
		Reward:     st.Reward,
		Status:     st.Status,
		Author:     brief(users[st.AuthorID]),
		Responses:  dtos,
		HasMine:    hasMine,
		MyResponse: myResponseID,
		CanSettle:  canSettle,
		CreatedAt:  st.CreatedAt.Format("2006-01-02 15:04"),
	}
}

// Create 发布求看：事务内冻结悬赏硬币
func (h *StoryHandler) Create(c *gin.Context) {
	user := middleware.CurrentUser(c)
	var req createStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题与悬赏硬币为必填，悬赏至少 1 枚"})
		return
	}

	if len(req.OldPhotos) == 0 || strings.TrimSpace(req.Memory) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请至少上传一张老照片并写下过去回忆"})
		return
	}

	story := &models.Story{
		AuthorID:  user.ID,
		Title:     strings.TrimSpace(req.Title),
		City:      strings.TrimSpace(req.City),
		Location:  strings.TrimSpace(req.Location),
		Memory:    strings.TrimSpace(req.Memory),
		OldPhotos: strings.Join(req.OldPhotos, ","),
		Wish:      strings.TrimSpace(req.Wish),
		Reward:    req.Reward,
		Status:    "open",
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(story).Error; err != nil {
			return err
		}
		return services.FreezeBounty(tx, user.ID, req.Reward, story.ID, services.OpFreeze, "发布求看，冻结悬赏硬币")
	})
	if err != nil {
		mapCoinError(c, err)
		return
	}

	var fresh models.User
	h.DB.First(&fresh, user.ID)
	c.JSON(http.StatusOK, gin.H{"story": h.singleDTOWithViewer(story.ID, &fresh), "balance": fresh.Balance, "frozen": fresh.Frozen})
}

// AppendReward 追加悬赏：继续从可用余额冻结
func (h *StoryHandler) AppendReward(c *gin.Context) {
	user := middleware.CurrentUser(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var req appendRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "追加数量至少为 1 枚"})
		return
	}

	var story models.Story
	if err := h.DB.First(&story, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "需求不存在"})
		return
	}
	if story.AuthorID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有发起人可以追加悬赏"})
		return
	}
	if story.Status != "open" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需求已结算，无法追加"})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := services.FreezeBounty(tx, user.ID, req.Amount, story.ID, services.OpAppend, "追加悬赏，追加冻结硬币"); err != nil {
			return err
		}
		return tx.Model(&models.Story{}).Where("id = ?", story.ID).
			UpdateColumn("reward", gorm.Expr("reward + ?", req.Amount)).Error
	})
	if err != nil {
		mapCoinError(c, err)
		return
	}

	var fresh models.User
	h.DB.First(&fresh, user.ID)
	c.JSON(http.StatusOK, gin.H{"story": h.singleDTOWithViewer(story.ID, &fresh), "balance": fresh.Balance, "frozen": fresh.Frozen})
}

// CreateResponse 替他去拍：上传现场新照与寄语
func (h *StoryHandler) CreateResponse(c *gin.Context) {
	user := middleware.CurrentUser(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var req createResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请写下给 TA 的寄语"})
		return
	}

	var story models.Story
	if err := h.DB.First(&story, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "需求不存在"})
		return
	}
	if story.AuthorID == user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能替自己去拍，把这条回忆留给别人吧"})
		return
	}
	if story.Status != "open" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "这条需求已被确认结算"})
		return
	}
	if len(req.NewPhotos) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请至少上传一张现场新照片"})
		return
	}

	var dup int64
	h.DB.Model(&models.Response{}).Where("story_id = ? AND user_id = ?", story.ID, user.ID).Count(&dup)
	if dup > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "你已经为这条需求拍过现场了"})
		return
	}

	resp := &models.Response{
		StoryID:   story.ID,
		UserID:    user.ID,
		Message:   strings.TrimSpace(req.Message),
		NewPhotos: strings.Join(req.NewPhotos, ","),
	}
	if err := h.DB.Create(resp).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "1062") {
			c.JSON(http.StatusConflict, gin.H{"error": "你已经为这条需求拍过现场了"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"story": h.singleDTOWithViewer(story.ID, user)})
}

// Accept 发起人确认：事务内结算冻结硬币给代看人
func (h *StoryHandler) Accept(c *gin.Context) {
	user := middleware.CurrentUser(c)
	storyID, _ := strconv.Atoi(c.Param("id"))
	respID, _ := strconv.Atoi(c.Param("rid"))

	var story models.Story
	if err := h.DB.First(&story, storyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "需求不存在"})
		return
	}
	if story.AuthorID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有发起人可以确认结算"})
		return
	}
	if story.Status != "open" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "这条需求已经结算过了"})
		return
	}

	var resp models.Response
	if err := h.DB.Where("id = ? AND story_id = ?", respID, story.ID).First(&resp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "回应不存在"})
		return
	}
	if resp.UserID == user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能确认自己的回应"})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// 事务内重新锁定故事行，防止并发重复结算
		var locked models.Story
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, story.ID).Error; err != nil {
			return err
		}
		if locked.Status != "open" {
			return errors.New("这条需求已经结算过了")
		}
		if err := services.SettleBounty(tx, story.AuthorID, resp.UserID, story.Reward, story.ID, resp.ID); err != nil {
			return err
		}
		if err := tx.Model(&locked).Updates(map[string]interface{}{
			"status":               "settled",
			"accepted_response_id": resp.ID,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&models.Response{}).Where("id = ?", resp.ID).Update("accepted", true).Error
	})
	if err != nil {
		mapCoinError(c, err)
		return
	}

	var author models.User
	h.DB.First(&author, user.ID)
	var helper models.User
	h.DB.First(&helper, resp.UserID)
	c.JSON(http.StatusOK, gin.H{
		"story":     h.singleDTOWithViewer(story.ID, &author),
		"balance":   author.Balance,
		"frozen":    author.Frozen,
		"helperGot": story.Reward,
		"helperId":  helper.ID,
	})
}

// List Feed 流：可选 mine=1 只看自己发布，responded=1 只看自己代看
func (h *StoryHandler) List(c *gin.Context) {
	q := h.DB.Model(&models.Story{})
	viewer := middleware.CurrentUser(c)

	switch c.Query("scope") {
	case "mine":
		if viewer == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		q = q.Where("author_id = ?", viewer.ID)
	case "responded":
		if viewer == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		var ids []uint
		h.DB.Model(&models.Response{}).Where("user_id = ?", viewer.ID).Pluck("story_id", &ids)
		if len(ids) == 0 {
			c.JSON(http.StatusOK, gin.H{"stories": []storyDTO{}})
			return
		}
		q = q.Where("id IN ?", ids)
	}

	if city := strings.TrimSpace(c.Query("city")); city != "" {
		q = q.Where("city = ?", city)
	}
	if st := c.Query("status"); st == "open" || st == "settled" {
		q = q.Where("status = ?", st)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	size := 20

	var total int64
	q.Count(&total)

	var stories []models.Story
	q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&stories)

	c.JSON(http.StatusOK, gin.H{
		"stories": h.batchDTO(stories, viewer),
		"total":   total,
		"page":    page,
	})
}

func (h *StoryHandler) batchDTO(stories []models.Story, viewer *models.User) []storyDTO {
	if len(stories) == 0 {
		return []storyDTO{}
	}
	ids := make([]uint, len(stories))
	for i, st := range stories {
		ids[i] = st.ID
	}
	var responses []models.Response
	h.DB.Where("story_id IN ?", ids).Order("created_at ASC").Find(&responses)
	byStory := map[uint][]models.Response{}
	for _, r := range responses {
		byStory[r.StoryID] = append(byStory[r.StoryID], r)
	}
	out := make([]storyDTO, 0, len(stories))
	for _, st := range stories {
		out = append(out, h.buildStoryDTO(st, byStory[st.ID], viewer))
	}
	return out
}

func (h *StoryHandler) singleDTOWithViewer(id uint, viewer *models.User) storyDTO {
	var st models.Story
	h.DB.First(&st, id)
	var responses []models.Response
	h.DB.Where("story_id = ?", id).Order("created_at ASC").Find(&responses)
	return h.buildStoryDTO(st, responses, viewer)
}

func mapCoinError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInsufficient):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "可用硬币不足，去帮别人看一眼赚些硬币吧"})
	case errors.Is(err, services.ErrInvalidAmount):
		c.JSON(http.StatusBadRequest, gin.H{"error": "硬币数量不合法"})
	case errors.Is(err, services.ErrFrozenShort):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "冻结硬币异常，无法结算，请联系管理员"})
	case errors.Is(err, services.ErrSelfSettle):
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能结算给自己"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
