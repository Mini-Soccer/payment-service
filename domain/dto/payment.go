package dto

import (
	"payment-service/constants"
	"time"

	"github.com/google/uuid"
)

type PaymentRequest struct {
	PaymentLink    string          `json:"paymentLink"`
	OrderID        string          `json:"orderID"`
	ExpiredAt      time.Time       `json:"expiredAt"`
	Amount         float64         `json:"amount"`
	Description    *string         `json:"description"`
	CustomerDetail *CustomerDetail `json:"customerDetail"`
	ItemDetails    []ItemDetail    `json:"itemDetails"`
}

type CustomerDetail struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type ItemDetail struct {
	ID       string  `json:"id"`
	Amount   float64 `json:"amount"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
}

type PaymentRequestParam struct {
	Page       int     `form:"page" validate:"required"`
	Limit      int     `form:"limit" validate:"required"`
	SortColumn *string `form:"sortColumn"`
	SortOrder  *string `form:"sortOrder"`
}

type UpdatePaymentRequest struct {
	TransactionID *string                  `json:"transactionId"`
	Status        *constants.PaymentStatus `json:"status"`
	PaidAt        *time.Time               `json:"paidAt"`
	VANumber      *string                  `json:"vaNumber"`
	Bank          *string                  `json:"bank"`
	InvoiceLink   *string                  `json:"invoiceLink,omitempty"`
	Acquirer      *string                  `json:"acquirer"`
}

type PaymentResponse struct {
	UUID          uuid.UUID                     `json:"uuid"`
	OrderID       uuid.UUID                     `json:"orderID"`
	Amount        float64                       `json:"amount"`
	Status        constants.PaymentStatusString `json:"status"`
	PaymentLink   string                        `json:"paymentLink"`
	InvoiceLink   *string                       `json:"invoiceLink,omitempty"`
	TransactionID *string                       `json:"transactionId,omitempty"`
	VANumber      *string                       `json:"vaNumber,omitempty"`
	Bank          *string                       `json:"bank,omitempty"`
	Acquirer      *string                       `json:"acquirer,omitempty"`
	Description   *string                       `json:"description"`
	PaidAt        *time.Time                    `json:"paidAt,omitempty"`
	ExpiredAt     *time.Time                    `json:"expiredAt"`
	CreatedAt     *time.Time                    `json:"createdAt"`
	UpdatedAt     *time.Time                    `json:"updatedAt"`
}

/*
- dipakai buat webhook payment gateway (Midtrans) yang ngirim status transaksi ke backend kamu
*/
type Webhook struct {
	VANumbers         []VANumber                    `json:"va_numbers"`
	TransactionTime   string                        `json:"transaction_time"` // Waktu transaksi dibuat
	TransactionStatus constants.PaymentStatusString `json:"transaction_status"`
	TransactionID     string                        `json:"transaction_id"` // ID transaksi dari payment gateway
	StatusMessage     string                        `json:"status_message"`
	StatusCode        string                        `json:"status_code"`     // Kode status dari gateway
	SignatureKey      string                        `json:"signature_key"`   // Digunakan untuk validasi webhook
	SettlementTime    string                        `json:"settlement_time"` // Waktu transaksi sukses (kalau sudah paid)
	PaymentType       string                        `json:"payment_type"`    // Jenis pembayaran (bank_transfer, qris, gopay, dll)
	PaymentAmount     []PaymentAmount               `json:"payment_amount"`
	OrderID           uuid.UUID                     `json:"order_id"`     // ID order internal kamu (UUID)
	MerchantID        string                        `json:"merchant_id"`  // ID merchant di payment gateway
	GrossAmount       string                        `json:"gross_amount"` // Total tagihan
	FraudStatus       string                        `json:"fraud_status"` // Status fraud check (accept, challenge, dll)
	Currency          string                        `json:"currency"`     // Mata uang (IDR)
	Acquirer          *string                       `json:"acquirer"`     // Pihak acquiring bank / e-wallet
}

type VANumber struct {
	VaNumber string `json:"va_number"` // Daftar Virtual Account (khusus bank transfer)
	Bank     string `json:"bank"`
}

type PaymentAmount struct {
	PaidAt *string `json:"paid_at"`
	Amount *string `json:"amount"`
}
