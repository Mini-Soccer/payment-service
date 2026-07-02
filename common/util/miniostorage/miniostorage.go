package util

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"payment-service/config"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioUtil struct {
	client     *minio.Client
	bucketName string
}

type IMinioUtil interface {
	UploadFile(ctx context.Context, path string, file []byte) (string, error)
	GetPresignedURL(path string, expiry time.Duration) (string, error)
}

func NewMinioUtil(client *minio.Client, bucket string) IMinioUtil {
	return &MinioUtil{
		client:     client,
		bucketName: bucket,
	}
}

func (m *MinioUtil) UploadFile(ctx context.Context, path string, file []byte) (string, error) {
	const timeout = 60 * time.Second

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := m.client.PutObject(
		ctx,
		m.bucketName,
		path,
		bytes.NewReader(file),
		int64(len(file)),
		minio.PutObjectOptions{
			ContentType: "image/jpeg",
		},
	)
	if err != nil {
		return "", err
	}

	// waktu pakai docker, ganti url host
	publicURL := config.Config.Minio.PublicURL
	url := fmt.Sprintf("%s/%s/%s", publicURL, m.bucketName, path)

	return url, nil
}

func (m *MinioUtil) GetPresignedURL(path string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)

	presignedURL, err := m.client.PresignedGetObject(
		context.Background(),
		m.bucketName,
		path,
		expiry,
		reqParams,
	)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}
