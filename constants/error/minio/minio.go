package error

import "errors"

var (
	ErrConnection = errors.New("Failed connect to storage")
	ErrBucket     = errors.New("Failed to get list bucket")
	ErrSetPolicy  = errors.New("Error set policy")
	ErrGetPolicy  = errors.New("Error get policy")
)

var MinioErrors = []error{
	ErrConnection,
	ErrBucket,
	ErrSetPolicy,
	ErrGetPolicy,
}
