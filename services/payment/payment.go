package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	clients "payment-service/clients/midtrans"
	"payment-service/common/miniostorage"
	"payment-service/common/util"
	config2 "payment-service/config"
	"payment-service/constants"
	errPayment "payment-service/constants/error/payment"
	"payment-service/controllers/kafka"
	"payment-service/domain/dto"
	"payment-service/domain/models"
	"payment-service/repositories"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PaymentService struct {
	repository   repositories.IRepositoryRegistry
	minioStorage miniostorage.IMinioStorage
	kafka        kafka.IKafkaRegistry
	midtrans     clients.IMidtransClient
}

type IPaymentService interface {
	GetAllWithPagination(context.Context, *dto.PaymentRequestParam) (*util.PaginationResult, error)
	GetByUUID(context.Context, string) (*dto.PaymentResponse, error)
	Create(context.Context, *dto.PaymentRequest) (*dto.PaymentResponse, error)
	Webhook(context.Context, *dto.Webhook) error
}

func NewPaymentService(
	repository repositories.IRepositoryRegistry,
	minioStorage miniostorage.IMinioStorage,
	kafka kafka.IKafkaRegistry,
	midtrans clients.IMidtransClient,
) IPaymentService {
	return &PaymentService{
		repository:   repository,
		minioStorage: minioStorage,
		kafka:        kafka,
		midtrans:     midtrans,
	}
}

func (p *PaymentService) GetAllWithPagination(
	ctx context.Context,
	param *dto.PaymentRequestParam,
) (*util.PaginationResult, error) {
	payments, total, err := p.repository.GetPayment().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}
	paymentResults := make([]dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentResults = append(paymentResults, dto.PaymentResponse{
			UUID:          payment.UUID,
			TransactionID: payment.TransactionID,
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			Status:        payment.Status.GetStatusString(),
			PaymentLink:   payment.PaymentLink,
			InvoiceLink:   payment.InvoiceLink,
			VANumber:      payment.VANumber,
			Bank:          payment.Bank,
			Description:   payment.Description,
			ExpiredAt:     payment.ExpiredAt,
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		})
	}

	paginationParam := util.PaginationParam{
		Page:  param.Page,
		Limit: param.Limit,
		Count: total,
		Data:  paymentResults,
	}

	response := util.GeneratePagination(paginationParam)
	return &response, nil
}

func (p *PaymentService) GetByUUID(ctx context.Context, uuid string) (*dto.PaymentResponse, error) {
	fmt.Println("uuid", uuid)
	payment, err := p.repository.GetPayment().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return &dto.PaymentResponse{
		UUID:          payment.UUID,
		TransactionID: payment.TransactionID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Status:        payment.Status.GetStatusString(),
		PaymentLink:   payment.PaymentLink,
		InvoiceLink:   payment.InvoiceLink,
		VANumber:      payment.VANumber,
		Bank:          payment.Bank,
		Description:   payment.Description,
		ExpiredAt:     payment.ExpiredAt,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}, nil
}

func (p *PaymentService) Create(ctx context.Context, req *dto.PaymentRequest) (*dto.PaymentResponse, error) {
	var (
		txErr, err error
		payment    *models.Payment
		response   *dto.PaymentResponse
		midtrans   *clients.MidtransData
	)

	err = p.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		// memastikan tidak expired
		if !req.ExpiredAt.After(time.Now()) {
			return errPayment.ErrExpireAtInvalid
		}

		midtrans, txErr = p.midtrans.CreatePaymentLink(req)
		if txErr != nil {
			return txErr
		}

		paymentRequest := &dto.PaymentRequest{
			OrderID:     req.OrderID,
			Amount:      req.Amount,
			Description: req.Description,
			ExpiredAt:   req.ExpiredAt,
			PaymentLink: midtrans.RedirectURL,
		}
		payment, txErr = p.repository.GetPayment().Create(ctx, tx, paymentRequest)
		if txErr != nil {
			return txErr
		}

		txErr = p.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentID: payment.ID,
			Status:    payment.Status.GetStatusString(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	response = &dto.PaymentResponse{
		UUID:        payment.UUID,
		OrderID:     payment.OrderID,
		Amount:      payment.Amount,
		Status:      payment.Status.GetStatusString(),
		PaymentLink: payment.PaymentLink,
		Description: payment.Description,
	}
	return response, nil
}

func (p *PaymentService) convertToIndonesianMonth(englishMonth string) string {
	monthMap := map[string]string{
		"January":   "Januari",
		"February":  "Februari",
		"March":     "Maret",
		"April":     "April",
		"May":       "Mei",
		"June":      "Juni",
		"July":      "Juli",
		"August":    "Agustus",
		"September": "September",
		"October":   "Oktober",
		"November":  "November",
		"December":  "Desember",
	}
	indonesianMonth, ok := monthMap[englishMonth]
	if !ok {
		return errors.New("month not found").Error()
	}
	return indonesianMonth
}

func (p *PaymentService) generatePDF(req *dto.InvoiceRequest) ([]byte, error) {
	htmlTemplatePath := "template/invoice.html"
	htmlTemplate, err := os.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	jsonData, _ := json.Marshal(req)
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	pdf, err := util.GeneratePDFFromHTML(string(htmlTemplate), data)
	if err != nil {
		return nil, err
	}

	return pdf, nil
}

func (p *PaymentService) uploadToMinio(ctx context.Context, invoiceNumber string, pdf []byte) (string, error) {
	invoiceNumberReplace := strings.ToLower(strings.ReplaceAll(invoiceNumber, "/", "-"))
	filename := fmt.Sprintf("%s.pdf", invoiceNumberReplace)
	url, err := p.minioStorage.UploadFile(ctx, filename, pdf)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (p *PaymentService) randomNumber() int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := random.Intn(900000) + 100000
	return number
}

func (p *PaymentService) mapTransactionStatusToEvent(status constants.PaymentStatusString) string {
	var paymentStatus string
	switch status {
	case constants.PendingString:
		paymentStatus = strings.ToUpper(constants.PendingString.String())
	case constants.SettlementString:
		paymentStatus = strings.ToUpper(constants.SettlementString.String())
	case constants.ExpireString:
		paymentStatus = strings.ToUpper(constants.ExpireString.String())
	}
	return paymentStatus
}

func (p *PaymentService) produceToKafka(
	req *dto.Webhook,
	payment *models.Payment,
	paidAt *time.Time,
) error {
	event := dto.KafkaEvent{
		Name: p.mapTransactionStatusToEvent(req.TransactionStatus),
	}

	metadata := dto.KafkaMetaData{
		Sender:    "payment-service",
		SendingAt: time.Now().Format(time.RFC3339),
	}

	body := dto.KafkaBody{
		Type: "JSON",
		Data: &dto.KafkaData{
			OrderID:   payment.OrderID,
			PaymentID: payment.UUID,
			Status:    req.TransactionStatus.String(),
			PaidAt:    paidAt,
			ExpiredAt: *payment.ExpiredAt,
		},
	}

	kafkaMessage := dto.KafkaMessage{
		Event:    event,
		Metadata: metadata,
		Body:     body,
	}

	topic := config2.Config.Kafka.Topic
	kafkaMessageJSON, _ := json.Marshal(kafkaMessage)

	producer, err := p.kafka.GetKafkaProducer()
	if err != nil {
		return err
	}

	err = producer.ProduceMessage(topic, payment.OrderID, kafkaMessageJSON)
	if err != nil {
		return err
	}
	return nil
}

/*
- Webhook() ini adalah handler untuk notifikasi asynchronous dari Midtrans
- midtrans akan hit webhook saat user memilih metode pembayaran (Karena memilih metode pembayaran = perubahan state transaksi.)
- fungsi utamanya:
- * validasi & update status payment
- * simpan history perubahan status
- * kalau sudah settlement, generate invoice + upload PDF
- * publish event ke kafka
- * Midtrans → kirim HTTP POST → endpoint ini → sistem kamu update state internal.
*/

/*
- alur midtrans sebenarnya:
- * kamu create transaction → status initial
- * user buka payment page
- * user pilih metode (VA,QRIS,e-wallet,dll)
- midtrans:
- * generate VA/QR/deeplink
- * set status jadi pending
- * kirim webhook ke merchant (service kita)
*/

func (p *PaymentService) Webhook(ctx context.Context, req *dto.Webhook) error {
	var (
		txErr, err         error
		paymentAfterUpdate *models.Payment
		paidAt             *time.Time
		invoiceLink        string
		pdf                []byte
	)

	/*
		- semua logic penting dibungkus trx DB supaya:
		- * update payment
		- * insert history
		- * generate invoice link
		- * kalau salah satu gagal, rollback semua
	*/
	err = p.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		// validasi payment exist
		_, txErr = p.repository.GetPayment().FindByOrderID(ctx, req.OrderID.String())
		if txErr != nil {
			return txErr
		}

		// tentukan paid_at hanya kalau settlement
		if req.TransactionStatus == constants.SettlementString {
			now := time.Now()
			paidAt = &now
		}

		status := req.TransactionStatus.GetStatusInt() // mapping status midtrans -> status internal
		/*
			- ambil va & bank
			- webhook ini diasumsikan payment type VA (Virtual Account) berbasis bank
			- data VA baru pasti ada setelah user pilih metode bayar
			- jadi di midtrans perlu diatur hanya pembayaran VA saja
		*/
		var vaNumber, bank string

		if len(req.VANumbers) > 0 {
			vaNumber = req.VANumbers[0].VaNumber
			bank = req.VANumbers[0].Bank
		}
		/*
			- update payment utama
			- payment di create di endpoint create
			- ini ini webhook:
			- * status berubah
			- * data payment disinkronkan dengan midtrans
			- penting: webhook adalah single source of truth, bukan response frontend.
			- untuk update payment juga sudah idempotent, karena status sama, data sama
		*/
		_, txErr = p.repository.GetPayment().Update(ctx, tx, req.OrderID.String(), &dto.UpdatePaymentRequest{
			TransactionID: &req.TransactionID,
			Status:        &status,
			PaidAt:        paidAt,
			VANumber:      &vaNumber,
			Bank:          &bank,
			Acquirer:      req.Acquirer, // keisi kalau va bank mandiri
		})
		if txErr != nil {
			return txErr
		}

		// ambil ulang payment setelah update
		paymentAfterUpdate, txErr = p.repository.GetPayment().FindByOrderID(ctx, req.OrderID.String())
		if txErr != nil {
			return txErr
		}

		/*
			- simpan payment history, datanya dari data payment terupdate
			- manfaat:
			- * audittrail
			- * debugging
			- * analytics
			- di create payment history belum idemptent, perlu unique key.
			- UNIQUE (payment_id, status)
		*/
		txErr = p.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentID: paymentAfterUpdate.ID,
			Status:    paymentAfterUpdate.Status.GetStatusString(),
		})

		/*
			- khusus req yang isinya settlement -> generate invoice
			- artinya invoice hanya untuk payment sukses
			- setelah generate PDF, upload ke GCS
			- lalu simpan link invoice ke payment
		*/
		if req.TransactionStatus == constants.SettlementString && paymentAfterUpdate.PaidAt == nil {
			paidDay := paidAt.Format("02")
			paidMonth := p.convertToIndonesianMonth(paidAt.Format("January"))
			paidYear := paidAt.Format("2006")
			invoiceNumber := fmt.Sprintf("INV/%s/ORD/%d", time.Now().Format(time.DateOnly), p.randomNumber())
			total := util.RupiahFormat(&paymentAfterUpdate.Amount)
			invoiceRequest := &dto.InvoiceRequest{
				InvoiceNumber: invoiceNumber,
				Data: dto.InvoiceData{
					PaymentDetail: dto.InvoicePaymentDetail{
						PaymentMethod: req.PaymentType,
						BankName:      strings.ToUpper(*paymentAfterUpdate.Bank),
						VANumber:      *paymentAfterUpdate.VANumber,
						Date:          fmt.Sprintf("%s %s %s", paidDay, paidMonth, paidYear),
						IsPaid:        true,
					},
					Items: []dto.InvoiceItem{
						{
							Description: *paymentAfterUpdate.Description,
							Price:       total,
						},
					},
					Total: total,
				},
			}
			pdf, txErr = p.generatePDF(invoiceRequest)
			if txErr != nil {
				return txErr
			}

			invoiceLink, txErr = p.uploadToMinio(ctx, invoiceNumber, pdf)
			if txErr != nil {
				return txErr
			}

			_, txErr = p.repository.GetPayment().Update(ctx, tx, req.OrderID.String(), &dto.UpdatePaymentRequest{
				InvoiceLink: &invoiceLink,
			})
			if txErr != nil {
				return txErr
			}
		}
		// kalau semua aman commit transaction DB
		return nil
	})
	if err != nil {
		return err
	}

	/*
		- publish event kafka diluar transaction
		- kalau kafka down, payment tetap sukses
	*/
	err = p.produceToKafka(req, paymentAfterUpdate, paidAt)
	if err != nil {
		return err
	}

	return nil
}
