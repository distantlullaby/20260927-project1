package middleware

import (
	"net/http"
	"strings"
	"time"

	"memorylink/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 为用户签发 JWT
func GenerateToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JWTTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.JWTSecret))
}

// parseToken 解析并校验 token
func parseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}

// Auth 必须登录
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader(config.TokenHeader)
		tokenStr = strings.TrimPrefix(tokenStr, config.TokenPrefix)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		claims, err := parseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已失效，请重新登录"})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// CurrentUserID 从上下文取当前用户 ID（未登录为 0）
func CurrentUserID(c *gin.Context) uint {
	if v, ok := c.Get("user_id"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}
