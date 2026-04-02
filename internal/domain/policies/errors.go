package policies

import "errors"

var (
	ErrPaymentAddressAllocationIssuedAtRequired = errors.New("issued at is required")
	ErrAddressPolicyIDRequired                  = errors.New("address policy id is required")
	ErrAddressPolicyChainMismatch               = errors.New("address policy chain mismatch")
	ErrAddressPolicyNotEnabled                  = errors.New("address policy is not enabled")
	ErrAddressPolicyPreviewNotSupported         = errors.New("address preview is not supported for this address policy")
	ErrExpectedAmountMinorInvalid               = errors.New("expected amount minor must be greater than zero")

	ErrPaymentReceiptStatusNotificationIDInvalid              = errors.New("notification id must be greater than zero")
	ErrPaymentReceiptStatusNotificationDeliveredAtRequired    = errors.New("delivered at is required")
	ErrPaymentReceiptStatusNotificationCurrentAttemptsInvalid = errors.New("current attempts must be greater than or equal to zero")
	ErrPaymentReceiptStatusNotificationMaxAttemptsInvalid     = errors.New("max attempts must be greater than zero")
	ErrPaymentReceiptStatusNotificationNowRequired            = errors.New("now is required")
	ErrPaymentReceiptStatusNotificationFailureReasonRequired  = errors.New("notification failure reason is required")
	ErrPaymentReceiptStatusNotificationRetryDelayInvalid      = errors.New("retry delay must be greater than zero")
)
