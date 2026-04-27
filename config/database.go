package config

import (
	"fmt"
	"log"
	"net/url"
	errWrap "payment-service/common/error"
	errConstant "payment-service/constants/error"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase() (*gorm.DB, error) {
	dbConfig := Config.Database
	encodedPassword := url.QueryEscape(dbConfig.Password)

	// 1. Connect TANPA database (default ke postgres)
	baseURI := fmt.Sprintf("postgresql://%s:%s@%s:%d/postgres?sslmode=disable",
		dbConfig.Username,
		encodedPassword,
		dbConfig.Host,
		dbConfig.Port,
	)

	baseDB, err := gorm.Open(postgres.Open(baseURI), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 2. Create database jika belum ada
	createDBQuery := fmt.Sprintf("CREATE DATABASE %s", dbConfig.Name)
	err = baseDB.Exec(createDBQuery).Error
	if err != nil {
		errWrap.WrapError(errConstant.ErrSQLError)
	} else {
		log.Println("database ready:", dbConfig.Name)
	}

	// 3. Connect ke database target
	targetURI := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		dbConfig.Username,
		encodedPassword,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)

	db, err := gorm.Open(postgres.Open(targetURI), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 4. Setup connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConnection)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConnection)
	sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.MaxLifeTimeConnection) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(dbConfig.MaxIdleTime) * time.Second)

	return db, nil
}
