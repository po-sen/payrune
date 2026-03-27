package policies

import "errors"

var (
	ErrPaymentAddressAllocationIssuedAtRequired = errors.New("issued at is required")

	ErrPaymentReceiptStatusNotificationIDInvalid              = errors.New("notification id must be greater than zero")
	ErrPaymentReceiptStatusNotificationDeliveredAtRequired    = errors.New("delivered at is required")
	ErrPaymentReceiptStatusNotificationCurrentAttemptsInvalid = errors.New("current attempts must be greater than or equal to zero")
	ErrPaymentReceiptStatusNotificationMaxAttemptsInvalid     = errors.New("max attempts must be greater than zero")
	ErrPaymentReceiptStatusNotificationNowRequired            = errors.New("now is required")
	ErrPaymentReceiptStatusNotificationLastErrorRequired      = errors.New("last error is required")
	ErrPaymentReceiptStatusNotificationRetryDelayInvalid      = errors.New("retry delay must be greater than zero")
)
