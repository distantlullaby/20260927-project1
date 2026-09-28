package handlers

import (
	"net/http"

	"memorylink/config"
	"memorylink/middleware"
	"memorylink/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type registerReq struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Nickname string `json:"nickname" binding:"max=50"`
}

// Register 注册并赠送初始记忆硬币（事务：建用户 + 写赠送流水）
func Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名 2-50 位，密码至少 6 位"})
		return
	}
	nick := req.Nickname
	if nick == "" {
		nick = req.Username
	}

	user := models.User{
		Username:     req.Username,
		PasswordHash: hashPassword(req.Password),
		Nickname:     nick,
		Balance:      config.RegisterGiftCoins,
	}

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Create(&models.CoinLedger{
			UserID:       user.ID,
			Change:       config.RegisterGiftCoins,
			Direction:    models.CoinDirectionIn,
			BalanceAfter: user.Balance,
			Type:         models.CoinTypeRegister,
			Remark:       "注册赠送初始硬币",
		}).Error
	})
	if err != nil {
		if isDuplicateKey(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已被注册"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	token, _ := middleware.GenerateToken(user.ID, user.Username)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录
func Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
		return
	}
	var user models.User
	if err := models.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	token, _ := middleware.GenerateToken(user.ID, user.Username)
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// Me 当前登录用户信息
func Me(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var user models.User
	if err := models.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func hashPassword(pwd string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(h)
}
