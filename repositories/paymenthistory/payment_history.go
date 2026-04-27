package repositories

import (
	"context"
	errWrap "payment-service/common/error"
	errConstant "payment-service/constants/error"
	"payment-service/domain/dto"
	"payment-service/domain/models"

	"gorm.io/gorm"
)

type PaymentHistoryRepository struct {
	db *gorm.DB
}

type IPaymentHistoryRepository interface {
	Create(context.Context, *gorm.DB, *dto.PaymentHistoryRequest) error
}

func NewPaymentHistoryRepository(db *gorm.DB) IPaymentHistoryRepository {
	return &PaymentHistoryRepository{db: db}
}

func (p *PaymentHistoryRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	req *dto.PaymentHistoryRequest,
) error {
	paymentHistory := models.PaymentHistory{
		PaymentID: req.PaymentID,
		Status:    req.Status,
	}

	err := tx.
		WithContext(ctx).
		Create(&paymentHistory).
		Error
	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}
	return nil
}
