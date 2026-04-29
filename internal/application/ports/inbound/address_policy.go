package inbound

import (
	"context"
	"time"
)

type AddressPolicy struct {
	AddressPolicyID string
	Chain           string
	Network         string
	Scheme          string
	AssetReference  string
	Decimals        uint8
	Enabled         bool
}

type ListAddressPoliciesResponse struct {
	Chain           string
	AddressPolicies []AddressPolicy
}

type AllocatePaymentAddressInput struct {
	Chain               string
	AddressPolicyID     string
	ExpectedAmountMinor int64
	CustomerReference   string
	IdempotencyKey      string
}

type AllocatePaymentAddressResponse struct {
	PaymentAddressID    string
	AddressPolicyID     string
	ExpectedAmountMinor int64
	Chain               string
	Network             string
	Scheme              string
	AssetReference      string
	Decimals            uint8
	Address             string
	CustomerReference   string
	IdempotencyReplayed bool
}

type GetPaymentAddressStatusInput struct {
	Chain            string
	PaymentAddressID int64
}

type GetPaymentAddressStatusResponse struct {
	PaymentAddressID        string
	AddressPolicyID         string
	ExpectedAmountMinor     int64
	Chain                   string
	Network                 string
	Scheme                  string
	AssetReference          string
	Decimals                uint8
	Address                 string
	CustomerReference       string
	PaymentStatus           string
	ObservedTotalMinor      int64
	ConfirmedTotalMinor     int64
	UnconfirmedTotalMinor   int64
	RequiredConfirmations   int32
	LastObservedBlockHeight int64
	IssuedAt                time.Time
	FirstObservedAt         *time.Time
	PaidAt                  *time.Time
	ConfirmedAt             *time.Time
	ExpiresAt               *time.Time
	LastError               string
}

type ListAddressPoliciesUseCase interface {
	Execute(ctx context.Context, chain string) (ListAddressPoliciesResponse, error)
}

type AllocatePaymentAddressUseCase interface {
	Execute(
		ctx context.Context,
		input AllocatePaymentAddressInput,
	) (AllocatePaymentAddressResponse, error)
}

type GetPaymentAddressStatusUseCase interface {
	Execute(
		ctx context.Context,
		input GetPaymentAddressStatusInput,
	) (GetPaymentAddressStatusResponse, error)
}
