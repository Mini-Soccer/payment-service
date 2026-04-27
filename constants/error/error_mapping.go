package error

import (
	errMinio "payment-service/constants/error/minio"
	errPayment "payment-service/constants/error/payment"
)

func ErrMapping(err error) bool {
	var (
		GeneralErrors = GeneralErrors
		TimeErrors    = errPayment.PaymentErrors
		MinioErrors   = errMinio.MinioErrors
	)

	allErrors := make([]error, 0)
	allErrors = append(allErrors, GeneralErrors...)
	allErrors = append(allErrors, TimeErrors...)
	allErrors = append(allErrors, MinioErrors...)

	for _, item := range allErrors {
		if err.Error() == item.Error() {
			return true
		}
	}

	return false
}
