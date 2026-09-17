package main

import (
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func initDB() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		slog.Error("DB_DSN 未设置")
		os.Exit(1)
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		slog.Error("数据库连接失败", "错误", err)
		os.Exit(1)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		slog.Error("获取数据库实例失败", "错误", err)
		os.Exit(1)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := DB.AutoMigrate(&Image{}); err != nil {
		slog.Error("数据库迁移失败", "错误", err)
		os.Exit(1)
	}

	slog.Info("数据库已连接")
}
