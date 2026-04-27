package clients

import (
	"math"
	errConstant "payment-service/constants/error/payment"
	"payment-service/domain/dto"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/sirupsen/logrus"
)

type MidtransClient struct {
	ServerKey    string // sandbox / production
	IsProduction bool   // flag sanbox / production
}

type IMidtransClient interface {
	// mengirim request ke midtrans snap
	// nanti menghasilkan payment link & token
	CreatePaymentLink(request *dto.PaymentRequest) (*MidtransData, error)
}

func NewMidtransClient(serverKey string, isProduction bool) *MidtransClient {
	return &MidtransClient{
		ServerKey:    serverKey,
		IsProduction: isProduction,
	}
}

func (c *MidtransClient) CreatePaymentLink(request *dto.PaymentRequest) (*MidtransData, error) {
	var (
		snapClient     snap.Client
		isProduction   = midtrans.Sandbox // flag default nya development
		expiryUnit     string
		expiryDuration int64
	)

	expiryDateTime := request.ExpiredAt
	currentTime := time.Now()
	duration := expiryDateTime.Sub(currentTime) // mendapatkan selisih waktu sekarang vs expired
	if duration <= 0 {                          // kalau expired sudah lewat -> reject (expired tidak boleh di masa lalu)
		logrus.Errorf("Expired at is invalid")
		return nil, errConstant.ErrExpireAtInvalid
	}

	// menentukan berlakunya payment link (menghitung expiry otomatis)
	// math.Ceil dipakai supaya payment link TIDAK PERNAH expired lebih cepat dari waktu ExpiredAt yang di tentukan (pembulatan keatas)
	if duration.Hours() >= 24 {
		expiryUnit = "day"
		expiryDuration = int64(math.Ceil(duration.Hours() / 24))
	} else if duration.Hours() >= 1 {
		expiryUnit = "hour"
		expiryDuration = int64(math.Ceil(duration.Hours()))
	} else {
		expiryUnit = "minute"
		expiryDuration = int64(math.Ceil(duration.Minutes()))
	}

	snapClient.New(c.ServerKey, isProduction) // set server key sanbox / production
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  request.OrderID,       // id unik transaksi
			GrossAmt: int64(request.Amount), // total pembayaran
		},
		CustomerDetail: &midtrans.CustomerDetails{ // dipakai midtrans buat invoice, email notifikasi, fraud check
			FName: request.CustomerDetail.Name,
			Email: request.CustomerDetail.Email,
			Phone: request.CustomerDetail.Phone,
		},
		Items: &[]midtrans.ItemDetails{ // disini hanya ambil item index 0, kalau multi item -> ini perlu di loop
			{
				ID:    request.ItemDetails[0].ID,
				Price: int64(request.ItemDetails[0].Amount),
				Qty:   int32(request.ItemDetails[0].Quantity),
				Name:  request.ItemDetails[0].Name,
			},
		},
		Expiry: &snap.ExpiryDetails{ // membuat payment link expired otomatis
			Unit:     expiryUnit,
			Duration: expiryDuration,
		},
	}

	// buat transaksi ke midtrans
	response, err := snapClient.CreateTransaction(req)
	if err != nil {
		logrus.Errorf("Error create transaction: %v", err)
		return nil, err
	}

	// midtrans return redirect url & token
	return &MidtransData{
		RedirectURL: response.RedirectURL,
		Token:       response.Token,
	}, nil
}
