package main

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"
)

func SaveImageLogic(file multipart.File, header *multipart.FileHeader) (*Image, error) {
	cosURL, err := UploadToCOS(file, header)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	image := Image{
		Name: header.Filename,
		Url:  cosURL,
	}

	if err := DB.Create(&image).Error; err != nil {
		return nil, fmt.Errorf("数据库入库失败: %w", err)
	}

	slog.Info("图片已保存", "名称", image.Name)
	return &image, nil
}

func GetImageListLogic(pageNo, pageSize string) ([]Image, int64, error) {
	var images []Image
	var total int64

	pn := 1
	ps := 20
	if v, err := strconv.Atoi(pageNo); err == nil && v > 0 {
		pn = v
	}
	if v, err := strconv.Atoi(pageSize); err == nil && v > 0 {
		ps = v
	}

	query := DB.Model(&Image{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pn - 1) * ps
	err := query.Offset(offset).Limit(ps).Order("id desc").Find(&images).Error
	return images, total, err
}

func DeleteImageLogic(id string) error {
	var image Image
	if err := DB.Where("id = ?", id).First(&image).Error; err != nil {
		return fmt.Errorf("图片不存在")
	}

	if err := DB.Delete(&image).Error; err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}

	if cosClientObj != nil && image.Url != "" {
		u, err := url.Parse(image.Url)
		if err == nil {
			objectKey := strings.TrimPrefix(u.Path, "/")
			if _, err := cosClientObj.Object.Delete(context.Background(), objectKey); err != nil {
				slog.Error("COS 文件删除失败", "objectKey", objectKey, "错误", err)
			}
		}
	}

	return nil
}
