package util

import (
	minioUtil "payment-service/common/util/miniostorage"
	ratelimiterUtil "payment-service/common/util/ratelimiter"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

type Registry struct {
	client     *minio.Client
	bucketName string
	redis      *redis.Client
	maxRequest int64
	duration   time.Duration
}

type IUtilRegistry interface {
	GetMinioUtil() minioUtil.IMinioUtil
	GetRatelimiterUtil() ratelimiterUtil.IRatelimiterUtil
}

func NewUtilRegistry(client *minio.Client, bucketName string, redis *redis.Client, maxRequest int64, duration time.Duration) IUtilRegistry {
	return &Registry{
		client:     client,
		bucketName: bucketName,
		redis:      redis,
		maxRequest: maxRequest,
		duration:   duration,
	}
}

func (r *Registry) GetMinioUtil() minioUtil.IMinioUtil {
	return minioUtil.NewMinioUtil(r.client, r.bucketName)
}

func (r *Registry) GetRatelimiterUtil() ratelimiterUtil.IRatelimiterUtil {
	return ratelimiterUtil.NewRatelimiterUtil(r.redis, r.maxRequest, r.duration)
}
