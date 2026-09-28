package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"memorylink/config"
	"memorylink/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// Upload 上传单张图片，返回可访问 URL
func Upload(c *gin.Context) {
	_ = middleware.CurrentUserID(c) // 登录校验由路由中间件保证

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
		return
	}
	if file.Size > config.MaxUpload {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片不能超过 10MB"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 jpg/png/gif/webp 格式"})
		return
	}

	name := uuid.NewString() + ext
	dst := filepath.Join(config.UploadDir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存失败: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + name})
}
