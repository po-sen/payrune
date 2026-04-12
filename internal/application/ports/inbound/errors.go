package inbound

import "errors"

var (
	// Business / user-facing application errors.
	ErrChainNotSupported       = errors.New("chain is not supported")
	ErrInvalidAddressPolicyID  = errors.New("address policy id is invalid")
	ErrAddressPolicyNotFound   = errors.New("address policy is not supported")
	ErrAddressPolicyNotEnabled = errors.New("address policy is not enabled")
	ErrAddressPoolExhausted    = errors.New("address pool is exhausted")
	ErrInvalidExpectedAmount   = errors.New("expected amount is invalid")
	ErrIdempotencyKeyConflict  = errors.New("idempotency key conflicts with existing payment address")
	ErrPaymentAddressNotFound  = errors.New("payment address is not found")
	ErrDependencyFailure       = errors.New("dependency failure")
	ErrInternalFailure         = errors.New("internal failure")

	// Use case dependency / configuration errors.
	ErrUnitOfWorkNotConfigured                     = errors.New("unit of work is not configured")
	ErrIssuedPaymentAddressDeriverNotConfigured    = errors.New("issued payment address deriver is not configured")
	ErrAddressPolicyReaderNotConfigured            = errors.New("address policy reader is not configured")
	ErrClockNotConfigured                          = errors.New("clock is not configured")
	ErrPaymentAddressStatusFinderNotConfigured     = errors.New("payment address status finder is not configured")
	ErrBlockchainReceiptObserverNotConfigured      = errors.New("blockchain receipt observer is not configured")
	ErrPaymentReceiptStatusNotifierNotConfigured   = errors.New("payment receipt status notifier is not configured")
	ErrPaymentAddressAllocationStoreNotConfigured  = errors.New("payment address allocation store is not configured")
	ErrPaymentAddressIdempotencyStoreNotConfigured = errors.New("payment address idempotency store is not configured")
	ErrPaymentReceiptTrackingStoreNotConfigured    = errors.New("payment receipt tracking store is not configured")
	ErrPaymentReceiptStatusOutboxNotConfigured     = errors.New("payment receipt status notification outbox is not configured")

	// Application validation / consistency errors.
	ErrBatchSizeMustBeGreaterThanZero                      = errors.New("batch size must be greater than zero")
	ErrMaxAttemptsMustBeGreaterThanZero                    = errors.New("max attempts must be greater than zero")
	ErrRetryDelayMustBeGreaterThanZero                     = errors.New("retry delay must be greater than zero")
	ErrPollChainRequiredWhenPollNetworkSet                 = errors.New("poll chain is required when poll network is set")
	ErrPaymentAddressPolicyNotConfigured                   = errors.New("payment address policy is not configured")
	ErrPaymentAddressIDMustBeGreaterThanZero               = errors.New("payment address id must be greater than zero")
	ErrPaymentAddressIdempotencyRecordIncomplete           = errors.New("payment address idempotency record is incomplete")
	ErrCompletedIdempotencyRecordMissingIssuedAllocation   = errors.New("completed payment address idempotency record references missing issued allocation")
	ErrIdempotencyClaimConflictWithoutCompletedRecord      = errors.New("idempotency key claim conflict occurred but no completed idempotency record was found")
	ErrPaymentAddressAllocationReservationAttemptInvalid   = errors.New("payment address allocation reservation attempt is invalid")
	ErrPaymentAddressAllocationReservationAttemptsRequired = errors.New("payment address allocation reservation attempts are required")
)
