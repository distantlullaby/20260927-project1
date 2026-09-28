package config

import (
	"os"
	"path/filepath"
)

// Config 运行期配置，全部支持通过环境变量覆盖
type Config struct {
	ServerPort string
	DSN        string // 不指定库名的 DSN，用于建库
	DBDSN      string // 指定库名的 DSN
	DBName     string
	UploadDir  string
	InitCoins  int
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() *Config {
	user := getenv("DB_USER", "root")
	pass := getenv("DB_PASSWORD", "123456") // 默认本地开发密码，可用环境变量覆盖
	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "3306")
	dbName := getenv("DB_NAME", "memory_connect")

	baseDSN := user + ":" + pass + "@tcp(" + host + ":" + port + ")/?charset=utf8mb4&parseTime=True&loc=Local"
	dbDSN := user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"

	uploadDir := getenv("UPLOAD_DIR", "./uploads")
	if abs, err := filepath.Abs(uploadDir); err == nil {
		uploadDir = abs
	}

	return &Config{
		ServerPort: getenv("SERVER_PORT", "8091"),
		DSN:        baseDSN,
		DBDSN:      dbDSN,
		DBName:     dbName,
		UploadDir:  uploadDir,
		InitCoins:  100,
	}
}
