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
	bucketURL := os.Getenv("cosBucketURL")
	if bucketURL == "" {
		slog.Error("cosBucketURL 未设置")
		os.Exit(1)
	}
	secretID := os.Getenv("cosSecretID")
	if secretID == "" {
		slog.Error("cosSecretID 未设置")
		os.Exit(1)
	}
	secretKey := os.Getenv("cosSecretKey")
	if secretKey == "" {
		slog.Error("cosSecretKey 未设置")
		os.Exit(1)
	}

	u, err := url.Parse(bucketURL)
	if err != nil {
		slog.Error("cosBucketURL 解析失败", "错误", err)
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
	cosPathPrefix := os.Getenv("cosPathPrefix")
	objectKey := fmt.Sprintf("%s%d%s", cosPathPrefix, time.Now().UnixNano(), ext)

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

	baseURL := os.Getenv("cosBucketURL")
	if cdn := os.Getenv("cosCDNURL"); cdn != "" {
		baseURL = cdn
	}
	return baseURL + "/" + objectKey, nil
}
