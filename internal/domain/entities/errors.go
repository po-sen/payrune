package entities

import "errors"

var (
	ErrPaymentAddressIDInvalid             = errors.New("payment address id must be greater than zero")
	ErrAddressPolicyIDRequired             = errors.New("address policy id is required")
	ErrAddressPolicyChainMismatch          = errors.New("address policy chain mismatch")
	ErrAddressPolicyNotEnabled             = errors.New("address policy is not enabled")
	ErrExpectedAmountMinorInvalid          = errors.New("expected amount minor must be greater than zero")
	ErrChainInvalid                        = errors.New("chain is invalid")
	ErrNetworkInvalid                      = errors.New("network is invalid")
	ErrAddressRequired                     = errors.New("address is required")
	ErrIssuedAtRequired                    = errors.New("issued at is required")
	ErrRequiredConfirmationsInvalid        = errors.New("required confirmations must be greater than zero")
	ErrObservedAtRequired                  = errors.New("observed time is required")
	ErrPaymentReceiptFailureReasonRequired = errors.New("payment receipt failure reason is required")
	ErrAddressPolicyMismatch               = errors.New("address policy mismatch")
	ErrDerivationFailureReasonRequired     = errors.New("derivation failure reason is required")
	ErrPaymentAddressAllocationNotIssued   = errors.New("payment address allocation is not issued")
	ErrExpiresAtRequired                   = errors.New("expires at is required")
)
