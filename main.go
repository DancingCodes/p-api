package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	_ = godotenv.Load()

	initCOS()
	initDB()
	r := setupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("服务已启动(Port:" + port + ")")
	if err := r.Run(":" + port); err != nil {
		slog.Error("服务启动失败", "error", err)
		os.Exit(1)
	}
}
