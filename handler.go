package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		api.POST("/admin/verify", VerifyAdmin)

		api.GET("/image/list", GetImageList)
		api.GET("/image/detail", GetImageDetail)
		api.POST("/image/upload", AdminAuth(), UploadImage)
		api.POST("/image/update", AdminAuth(), UpdateImage)
		api.DELETE("/image/delete", AdminAuth(), DeleteImage)

		api.GET("/video/list", GetVideoList)
		api.GET("/video/detail", GetVideoDetail)
		api.POST("/video/upload", AdminAuth(), UploadVideo)
		api.POST("/video/update", AdminAuth(), UpdateVideo)
		api.DELETE("/video/delete", AdminAuth(), DeleteVideo)
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
	category := c.DefaultQuery("category", "all")

	list, total, err := GetImageListLogic(pageNo, pageSize, category)
	if err != nil {
		slog.Error("获取图片列表失败", "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

func GetVideoList(c *gin.Context) {
	pageNo := c.DefaultQuery("pageNo", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	category := c.DefaultQuery("category", "all")

	list, total, err := GetVideoListLogic(pageNo, pageSize, category)
	if err != nil {
		slog.Error("获取视频列表失败", "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

func GetImageDetail(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		Error(c, "请传入图片 id")
		return
	}

	image, err := GetImageByIDLogic(id)
	if err != nil {
		Error(c, err.Error())
		return
	}
	Success(c, image)
}

func GetVideoDetail(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		Error(c, "请传入视频 id")
		return
	}

	video, err := GetVideoByIDLogic(id)
	if err != nil {
		Error(c, err.Error())
		return
	}
	Success(c, video)
}

func UploadImage(c *gin.Context) {
	name := strings.TrimSpace(c.PostForm("name"))
	category := strings.TrimSpace(c.PostForm("category"))
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, "请选择图片")
		return
	}
	defer file.Close()

	image, err := SaveImageLogic(file, header, name, category)
	if err != nil {
		slog.Error("上传图片失败", "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, image)
}

func UploadVideo(c *gin.Context) {
	name := strings.TrimSpace(c.PostForm("name"))
	category := strings.TrimSpace(c.PostForm("category"))

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, "请选择视频")
		return
	}
	defer file.Close()

	cover, coverHeader, err := c.Request.FormFile("cover")
	if err != nil {
		Error(c, "请上传视频封面")
		return
	}
	defer cover.Close()

	video, err := SaveVideoLogic(file, header, cover, coverHeader, name, category)
	if err != nil {
		slog.Error("上传视频失败", "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, video)
}

func UpdateImage(c *gin.Context) {
	id := strings.TrimSpace(c.PostForm("id"))
	name := strings.TrimSpace(c.PostForm("name"))
	category := strings.TrimSpace(c.PostForm("category"))
	if id == "" {
		Error(c, "请传入图片 id")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		file, header = nil, nil
	} else {
		defer file.Close()
	}

	image, err := UpdateImageLogic(id, name, category, file, header)
	if err != nil {
		slog.Error("更新图片失败", "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, image)
}

func UpdateVideo(c *gin.Context) {
	id := strings.TrimSpace(c.PostForm("id"))
	name := strings.TrimSpace(c.PostForm("name"))
	category := strings.TrimSpace(c.PostForm("category"))
	if id == "" {
		Error(c, "请传入视频 id")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		file, header = nil, nil
	} else {
		defer file.Close()
	}

	cover, coverHeader, err := c.Request.FormFile("cover")
	if err != nil {
		cover, coverHeader = nil, nil
	} else {
		defer cover.Close()
	}

	video, err := UpdateVideoLogic(id, name, category, file, header, cover, coverHeader)
	if err != nil {
		slog.Error("更新视频失败", "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, video)
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

func DeleteVideo(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		Error(c, "请传入视频 id")
		return
	}

	if err := DeleteVideoLogic(id); err != nil {
		slog.Error("删除视频失败", "id", id, "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, nil)
}
