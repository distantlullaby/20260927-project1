package middleware

import (
	"strings"

	"memoryconnect/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const ctxUserKey = "currentUser"

// extractToken 依次从 Authorization 头与 query 参数读取 token
func extractToken(c *gin.Context) string {
	if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return c.Query("token")
}

func loadUser(c *gin.Context, db *gorm.DB) *models.User {
	token := extractToken(c)
	if token == "" {
		return nil
	}
	var u models.User
	if err := db.Where("token = ?", token).First(&u).Error; err != nil {
		return nil
	}
	return &u
}

// RequiredAuth 必须登录，失败返回 401
func RequiredAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := loadUser(c, db)
		if u == nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "请先登录"})
			return
		}
		c.Set(ctxUserKey, u)
		c.Next()
	}
}

// OptionalAuth 有合法 token 就挂上用户，没有也放行（Feed 流游客可看）
func OptionalAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if u := loadUser(c, db); u != nil {
			c.Set(ctxUserKey, u)
		}
		c.Next()
	}
}

// CurrentUser 取出当前用户，未登录返回 nil
func CurrentUser(c *gin.Context) *models.User {
	if v, ok := c.Get(ctxUserKey); ok {
		if u, ok := v.(*models.User); ok {
			return u
		}
	}
	return nil
}

// CORS 允许本地前端跨域联调
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
