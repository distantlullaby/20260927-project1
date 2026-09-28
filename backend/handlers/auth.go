package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"memoryconnect/config"
	"memoryconnect/models"
	"memoryconnect/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

type authRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Nickname string `json:"nickname"`
}

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func userJSON(u *models.User) gin.H {
	nick := u.Nickname
	if nick == "" {
		nick = u.Username
	}
	return gin.H{
		"id":       u.ID,
		"username": u.Username,
		"nickname": nick,
		"avatar":   u.Avatar,
		"balance":  u.Balance,
		"frozen":   u.Frozen,
	}
}

// Register 注册即赠送初始记忆硬币
func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名至少 2 位，密码至少 6 位"})
		return
	}

	var existing int64
	h.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&existing)
	if existing > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已被注册"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	user := &models.User{
		Username: req.Username,
		Password: string(hash),
		Nickname: strings.TrimSpace(req.Nickname),
		Token:    newToken(),
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return services.GrantInitial(tx, user.ID, h.Cfg.InitCoins)
	})
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "1062") {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已被注册"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败：" + err.Error()})
		return
	}

	// 硬币在事务内由独立查询更新，回读以拿到最新 balance/frozen
	h.DB.First(user, user.ID)
	c.JSON(http.StatusOK, gin.H{"token": user.Token, "user": userJSON(user)})
}

// Login 登录并下发 token
func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}

	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	user.Token = newToken()
	if err := h.DB.Model(&user).Update("token", user.Token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": user.Token, "user": userJSON(&user)})
}
