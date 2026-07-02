package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/spf13/viper/remote"
)

var Config AppConfig

func Init() {
	Config = Load()
}

type AppConfig struct {
	Port                  int
	AppName               string
	AppEnv                string
	SignatureKey          string
	Database              Database
	Redis                 Redis
	RateLimiterMaxRequest float64
	RateLimiterTimeSecond int
	InternalService       InternalService
	Kafka                 Kafka
	Midtrans              Midtrans
	Minio                 Minio
}

type Database struct {
	Host                  string
	Port                  int
	Name                  string
	Username              string
	Password              string
	MaxOpenConnection     int
	MaxLifeTimeConnection int
	MaxIdleConnection     int
	MaxIdleTime           int
}

type Redis struct {
	Addr     string
	DB       int
	Password string
}

type Minio struct {
	Endpoint  string
	PublicURL string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type InternalService struct {
	User User
}

type User struct {
	Host         string
	SignatureKey string
}

type Kafka struct {
	Brokers     []string
	TimeoutInMS int
	MaxRetry    int
	Topic       string
}

type Midtrans struct {
	ServerKey    string
	ClientKey    string
	IsProduction bool
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("missing env: " + key)
	}
	return val
}

func Load() AppConfig {
	_ = godotenv.Load()

	port, _ := strconv.Atoi(mustEnv("PORT"))
	dbPort, _ := strconv.Atoi(mustEnv("DB_PORT"))
	dbMaxOpenConnection, _ := strconv.Atoi(mustEnv("DB_MAX_OPEN_CONNECTION"))
	dbMaxLifeTimeConnection, _ := strconv.Atoi(mustEnv("DB_MAX_LIFE_TIME_CONNECTION"))
	dbMaxIdleConnection, _ := strconv.Atoi(mustEnv("DB_MAX_IDLE_CONNECTION"))
	dbMaxIdleTime, _ := strconv.Atoi(mustEnv("DB_MAX_IDLE_TIME"))
	rateLimiterMaxRequest, _ := strconv.Atoi(mustEnv("RATE_LIMITER_MAX_REQUEST"))
	rateLimiterTimeSecond, _ := strconv.Atoi(mustEnv("RATE_LIMITER_TIME_SECOND"))
	minioUseSsl, _ := strconv.ParseBool(mustEnv("MINIO_USE_SSL"))
	kafkaTimeout, _ := strconv.Atoi(mustEnv("KAFKA_TIMEOUT_MS"))
	kafkaMaxRetry, _ := strconv.Atoi(mustEnv("KAFKA_MAX_RETRY"))
	kafkaBrokers := strings.Split(mustEnv("KAFKA_BROKERS"), ",")
	midtransIsProd, _ := strconv.ParseBool(mustEnv("MIDTRANS_IS_PRODUCTION"))
	redisDb, _ := strconv.Atoi(mustEnv("REDIS_DB"))

	return AppConfig{
		Port:         port,
		AppName:      mustEnv("APP_NAME"),
		AppEnv:       mustEnv("APP_ENV"),
		SignatureKey: mustEnv("SIGNATURE_KEY"),
		Database: Database{
			Host:                  mustEnv("DB_HOST"),
			Port:                  dbPort,
			Name:                  mustEnv("DB_NAME"),
			Username:              mustEnv("DB_USERNAME"),
			Password:              mustEnv("DB_PASSWORD"),
			MaxOpenConnection:     dbMaxOpenConnection,
			MaxLifeTimeConnection: dbMaxLifeTimeConnection,
			MaxIdleConnection:     dbMaxIdleConnection,
			MaxIdleTime:           dbMaxIdleTime,
		},
		Redis: Redis{
			Addr:     mustEnv("REDIS_ADDR"),
			DB:       redisDb,
			Password: mustEnv("REDIS_PASSWORD"),
		},
		RateLimiterMaxRequest: float64(rateLimiterMaxRequest),
		RateLimiterTimeSecond: rateLimiterTimeSecond,
		InternalService: InternalService{
			User: User{
				Host:         mustEnv("USER_SERVICE_HOST"),
				SignatureKey: mustEnv("USER_SERVICE_SIGNATURE_KEY"),
			},
		},
		Minio: Minio{
			Endpoint:  mustEnv("MINIO_ENDPOINT"),
			PublicURL: mustEnv("MINIO_PUBLIC_URL"),
			AccessKey: mustEnv("MINIO_ACCESS_KEY"),
			SecretKey: mustEnv("MINIO_SECRET_KEY"),
			Bucket:    mustEnv("MINIO_BUCKET"),
			UseSSL:    minioUseSsl,
		},
		Kafka: Kafka{
			Brokers:     kafkaBrokers,
			TimeoutInMS: kafkaTimeout,
			MaxRetry:    kafkaMaxRetry,
			Topic:       mustEnv("KAFKA_TOPICS"),
		},
		Midtrans: Midtrans{
			ServerKey:    mustEnv("MIDTRANS_SERVER_KEY"),
			ClientKey:    mustEnv("MIDTRANS_CLIENT_KEY"),
			IsProduction: midtransIsProd,
		},
	}
}
