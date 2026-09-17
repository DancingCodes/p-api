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

var imageCategories = map[string]struct{}{
	"poster": {},
	"banner": {},
	"detail": {},
	"promo":  {},
	"fold":   {},
}

var videoCategories = map[string]struct{}{
	"speech":   {},
	"showcase": {},
	"ai":       {},
}

func SaveImageLogic(file multipart.File, header *multipart.FileHeader, name, category string) (*Image, error) {
	name = strings.TrimSpace(name)
	category = strings.TrimSpace(category)
	if _, ok := imageCategories[category]; !ok {
		return nil, fmt.Errorf("图片分类无效")
	}

	cosURL, err := UploadToCOS(file, header)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	image := Image{
		Name:     name,
		Category: category,
		Url:      cosURL,
	}

	if err := DB.Create(&image).Error; err != nil {
		return nil, fmt.Errorf("数据库入库失败: %w", err)
	}

	slog.Info("图片已保存", "名称", image.Name, "分类", image.Category)
	return &image, nil
}

func SaveVideoLogic(
	file multipart.File,
	header *multipart.FileHeader,
	cover multipart.File,
	coverHeader *multipart.FileHeader,
	name, category string,
) (*Video, error) {
	name = strings.TrimSpace(name)
	category = strings.TrimSpace(category)
	if _, ok := videoCategories[category]; !ok {
		return nil, fmt.Errorf("视频分类无效")
	}
	if cover == nil || coverHeader == nil {
		return nil, fmt.Errorf("请上传视频封面")
	}

	cosURL, err := UploadToCOS(file, header)
	if err != nil {
		return nil, fmt.Errorf("上传视频失败: %w", err)
	}

	coverURL, err := UploadToCOS(cover, coverHeader)
	if err != nil {
		deleteCOSObject(cosURL)
		return nil, fmt.Errorf("上传封面失败: %w", err)
	}

	video := Video{
		Name:     name,
		Category: category,
		Url:      cosURL,
		CoverUrl: coverURL,
	}

	if err := DB.Create(&video).Error; err != nil {
		deleteCOSObject(cosURL)
		deleteCOSObject(coverURL)
		return nil, fmt.Errorf("数据库入库失败: %w", err)
	}

	slog.Info("视频已保存", "名称", video.Name, "分类", video.Category)
	return &video, nil
}

func GetImageByIDLogic(id string) (*Image, error) {
	var image Image
	if err := DB.Where("id = ?", id).First(&image).Error; err != nil {
		return nil, fmt.Errorf("图片不存在")
	}
	return &image, nil
}

func GetVideoByIDLogic(id string) (*Video, error) {
	var video Video
	if err := DB.Where("id = ?", id).First(&video).Error; err != nil {
		return nil, fmt.Errorf("视频不存在")
	}
	return &video, nil
}

func UpdateImageLogic(id, name, category string, file multipart.File, header *multipart.FileHeader) (*Image, error) {
	image, err := GetImageByIDLogic(id)
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	category = strings.TrimSpace(category)
	if _, ok := imageCategories[category]; !ok {
		return nil, fmt.Errorf("图片分类无效")
	}

	oldURL := image.Url
	if file != nil && header != nil {
		cosURL, err := UploadToCOS(file, header)
		if err != nil {
			return nil, fmt.Errorf("上传文件失败: %w", err)
		}
		image.Url = cosURL
	}

	image.Name = name
	image.Category = category

	if err := DB.Save(image).Error; err != nil {
		if image.Url != oldURL {
			deleteCOSObject(image.Url)
			image.Url = oldURL
		}
		return nil, fmt.Errorf("更新失败: %w", err)
	}

	if image.Url != oldURL {
		deleteCOSObject(oldURL)
	}

	slog.Info("图片已更新", "id", image.ID, "名称", image.Name, "分类", image.Category)
	return image, nil
}

func UpdateVideoLogic(
	id, name, category string,
	file multipart.File,
	header *multipart.FileHeader,
	cover multipart.File,
	coverHeader *multipart.FileHeader,
) (*Video, error) {
	video, err := GetVideoByIDLogic(id)
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	category = strings.TrimSpace(category)
	if _, ok := videoCategories[category]; !ok {
		return nil, fmt.Errorf("视频分类无效")
	}

	oldURL := video.Url
	oldCoverURL := video.CoverUrl
	newURL := oldURL
	newCoverURL := oldCoverURL

	if file != nil && header != nil {
		cosURL, err := UploadToCOS(file, header)
		if err != nil {
			return nil, fmt.Errorf("上传视频失败: %w", err)
		}
		newURL = cosURL
	}

	if cover != nil && coverHeader != nil {
		coverURL, err := UploadToCOS(cover, coverHeader)
		if err != nil {
			if newURL != oldURL {
				deleteCOSObject(newURL)
			}
			return nil, fmt.Errorf("上传封面失败: %w", err)
		}
		newCoverURL = coverURL
	}

	video.Name = name
	video.Category = category
	video.Url = newURL
	video.CoverUrl = newCoverURL

	if err := DB.Save(video).Error; err != nil {
		if newURL != oldURL {
			deleteCOSObject(newURL)
		}
		if newCoverURL != oldCoverURL {
			deleteCOSObject(newCoverURL)
		}
		return nil, fmt.Errorf("更新失败: %w", err)
	}

	if newURL != oldURL {
		deleteCOSObject(oldURL)
	}
	if newCoverURL != oldCoverURL {
		deleteCOSObject(oldCoverURL)
	}

	slog.Info("视频已更新", "id", video.ID, "名称", video.Name, "分类", video.Category)
	return video, nil
}

func GetImageListLogic(pageNo, pageSize, category string) ([]Image, int64, error) {
	var images []Image
	var total int64

	pn, ps := parsePage(pageNo, pageSize)
	query := DB.Model(&Image{})
	if category != "" && category != "all" {
		if _, ok := imageCategories[category]; !ok {
			return nil, 0, fmt.Errorf("图片分类无效")
		}
		query = query.Where("category = ?", category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pn - 1) * ps
	err := query.Offset(offset).Limit(ps).Order("id desc").Find(&images).Error
	return images, total, err
}

func GetVideoListLogic(pageNo, pageSize, category string) ([]Video, int64, error) {
	var videos []Video
	var total int64

	pn, ps := parsePage(pageNo, pageSize)
	query := DB.Model(&Video{})
	if category != "" && category != "all" {
		if _, ok := videoCategories[category]; !ok {
			return nil, 0, fmt.Errorf("视频分类无效")
		}
		query = query.Where("category = ?", category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pn - 1) * ps
	err := query.Offset(offset).Limit(ps).Order("id desc").Find(&videos).Error
	return videos, total, err
}

func DeleteImageLogic(id string) error {
	var image Image
	if err := DB.Where("id = ?", id).First(&image).Error; err != nil {
		return fmt.Errorf("图片不存在")
	}

	if err := DB.Delete(&image).Error; err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}

	deleteCOSObject(image.Url)
	return nil
}

func DeleteVideoLogic(id string) error {
	var video Video
	if err := DB.Where("id = ?", id).First(&video).Error; err != nil {
		return fmt.Errorf("视频不存在")
	}

	if err := DB.Delete(&video).Error; err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}

	deleteCOSObject(video.Url)
	deleteCOSObject(video.CoverUrl)
	return nil
}

func parsePage(pageNo, pageSize string) (int, int) {
	pn := 1
	ps := 20
	if v, err := strconv.Atoi(pageNo); err == nil && v > 0 {
		pn = v
	}
	if v, err := strconv.Atoi(pageSize); err == nil && v > 0 {
		ps = v
	}
	return pn, ps
}

func deleteCOSObject(fileURL string) {
	if cosClientObj == nil || fileURL == "" {
		return
	}
	u, err := url.Parse(fileURL)
	if err != nil {
		return
	}
	objectKey := strings.TrimPrefix(u.Path, "/")
	if objectKey == "" {
		return
	}
	if _, err := cosClientObj.Object.Delete(context.Background(), objectKey); err != nil {
		slog.Error("COS 文件删除失败", "objectKey", objectKey, "错误", err)
	}
}
