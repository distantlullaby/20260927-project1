package config

import "time"

const (
	ServerAddr = ":8080"

	// MySQL
	MySQLDSN = "root:123456@tcp(127.0.0.1:3306)/memorylink?charset=utf8mb4&parseTime=True&loc=Local"

	// JWT
	JWTSecret    = "memorylink-secret-2026"
	JWTTTL       = 7 * 24 * time.Hour
	TokenHeader  = "Authorization"
	TokenPrefix  = "Bearer "

	// 记忆硬币
	RegisterGiftCoins = 100 // 注册即赠送初始硬币

	// 文件上传
	UploadDir = "uploads"
	MaxUpload = 10 << 20 // 10MB
)
