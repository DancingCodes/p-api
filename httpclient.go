package main

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var cosClientObj *cos.Client

func initCOS() {
	bucketURL := os.Getenv("COS_BUCKET_URL")
	if bucketURL == "" {
		slog.Error("COS_BUCKET_URL 未设置")
		os.Exit(1)
	}
	secretID := os.Getenv("COS_SECRET_ID")
	if secretID == "" {
		slog.Error("COS_SECRET_ID 未设置")
		os.Exit(1)
	}
	secretKey := os.Getenv("COS_SECRET_KEY")
	if secretKey == "" {
		slog.Error("COS_SECRET_KEY 未设置")
		os.Exit(1)
	}

	u, err := url.Parse(bucketURL)
	if err != nil {
		slog.Error("COS_BUCKET_URL 解析失败", "错误", err)
		os.Exit(1)
	}
	cosClientObj = cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
}

func UploadToCOS(file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	CosPathPrefix := os.Getenv("COS_PATH_PREFIX")
	objectKey := fmt.Sprintf("%s%d%s", CosPathPrefix, time.Now().UnixNano(), ext)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentLength: header.Size,
		},
	}
	_, err := cosClientObj.Object.Put(ctx, objectKey, file, opt)
	if err != nil {
		return "", fmt.Errorf("COS 上传失败: %w", err)
	}

	baseURL := os.Getenv("COS_BUCKET_URL")
	if cdn := os.Getenv("COS_CDN_URL"); cdn != "" {
		baseURL = cdn
	}
	return baseURL + "/" + objectKey, nil
}
