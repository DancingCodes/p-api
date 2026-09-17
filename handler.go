package main

import (
	"log/slog"
	"mime/multipart"
	"os"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		api.POST("/admin/verify", VerifyAdmin)
		api.GET("/image/list", GetImageList)
		api.POST("/image/upload", AdminAuth(), UploadImage)
		api.DELETE("/image/delete", AdminAuth(), DeleteImage)
	}

	return r
}

func VerifyAdmin(c *gin.Context) {
	adminKey := os.Getenv("ADMIN_KEY")
	if adminKey == "" {
		Error(c, "未配置 ADMIN_KEY")
		return
	}

	key := c.GetHeader("X-Admin-Key")
	if key == adminKey {
		Success(c, nil)
	} else {
		Error(c, "密钥错误")
	}
}

func AdminAuth() gin.HandlerFunc {
	adminKey := os.Getenv("ADMIN_KEY")
	return func(c *gin.Context) {
		if c.GetHeader("X-Admin-Key") != adminKey || adminKey == "" {
			Error(c, "无权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetImageList(c *gin.Context) {
	pageNo := c.DefaultQuery("pageNo", "1")
	pageSize := c.DefaultQuery("pageSize", "20")

	list, total, err := GetImageListLogic(pageNo, pageSize)
	if err != nil {
		slog.Error("获取图片列表失败", "错误", err)
		Error(c, "获取列表失败")
		return
	}

	Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

func UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, "请选择图片")
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	image, err := SaveImageLogic(file, header)
	if err != nil {
		slog.Error("上传图片失败", "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, image)
}

func DeleteImage(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		Error(c, "请传入图片 id")
		return
	}

	if err := DeleteImageLogic(id); err != nil {
		slog.Error("删除图片失败", "id", id, "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, nil)
}
