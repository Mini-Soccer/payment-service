package miniostorage

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"payment-service/config"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	Client     *minio.Client
	BucketName string
}

type IMinioStorage interface {
	UploadFile(ctx context.Context, path string, file []byte) (string, error)
	GetPresignedURL(path string, expiry time.Duration) (string, error)
}

func NewMinioStorage(client *minio.Client, bucket string) IMinioStorage {
	return &MinioStorage{
		Client:     client,
		BucketName: bucket,
	}
}

func (m *MinioStorage) UploadFile(ctx context.Context, path string, file []byte) (string, error) {
	const timeout = 60 * time.Second
	contentType := http.DetectContentType(file)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := m.Client.PutObject(
		ctx,
		m.BucketName,
		path,
		bytes.NewReader(file),
		int64(len(file)),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", err
	}

	publicURL := config.Config.Minio.PublicURL
	url := fmt.Sprintf("%s/%s/%s", publicURL, m.BucketName, path)

	return url, nil
}

func (m *MinioStorage) GetPresignedURL(path string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)

	presignedURL, err := m.Client.PresignedGetObject(
		context.Background(),
		m.BucketName,
		path,
		expiry,
		reqParams,
	)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}
