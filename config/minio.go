package config

import (
	"context"
	"fmt"

	errWrap "payment-service/common/error"

	errMinio "payment-service/constants/error/minio"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinio() (*minio.Client, error) {
	config := Config

	minioClient, err := minio.New(config.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Minio.AccessKey, config.Minio.SecretKey, ""),
		Secure: config.Minio.UseSSL,
	})

	if err != nil {
		return nil, errWrap.WrapError(errMinio.ErrConnection)
	}

	ctx := context.Background()

	// test koneksi
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		return nil, errWrap.WrapError(errMinio.ErrBucket)
	}

	// nama bucket dari config
	bucketName := config.Minio.Bucket

	// cek bucket sudah ada atau belum
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, errWrap.WrapError(errMinio.ErrBucket)
	}

	// kalau belum ada → buat bucket
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, errWrap.WrapError(errMinio.ErrBucket)
		}
	}

	// cek existing policy
	existingPolicy, err := minioClient.GetBucketPolicy(ctx, bucketName)
	if err != nil {
		// kalau error selain "no policy", tetap fail
		return nil, errWrap.WrapError(errMinio.ErrGetPolicy)
	}

	// kalau belum ada policy → set
	// - semua orang boleh download / lihat file
	// - tidak boleh upload, delete, overwrite
	if existingPolicy == "" {
		policy := fmt.Sprintf(`{
			"Version":"2012-10-17",
			"Statement":[
				{
					"Effect":"Allow",
					"Principal":"*",
					"Action":["s3:GetObject"],
					"Resource":["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucketName)

		if err := minioClient.SetBucketPolicy(ctx, bucketName, policy); err != nil {
			return nil, errWrap.WrapError(errMinio.ErrSetPolicy)
		}
	}

	return minioClient, nil
}
