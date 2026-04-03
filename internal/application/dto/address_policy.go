package dto

import (
	"time"

	"payrune/internal/domain/valueobjects"
)

type AddressPolicy struct {
	AddressPolicyID string
	Chain           string
	Network         string
	Scheme          string
	MinorUnit       string
	Decimals        uint8
	Enabled         bool
}

type ListAddressPoliciesResponse struct {
	Chain           string
	AddressPolicies []AddressPolicy
}

type GenerateAddressInput struct {
	Chain           valueobjects.SupportedChain
	AddressPolicyID string
	Index           uint32
}

type GenerateAddressResponse struct {
	AddressPolicyID string
	Chain           string
	Network         string
	Scheme          string
	MinorUnit       string
	Decimals        uint8
	Index           uint32
	Address         string
}

type AllocatePaymentAddressInput struct {
	Chain               valueobjects.SupportedChain
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
	MinorUnit           string
	Decimals            uint8
	Address             string
	CustomerReference   string
	IdempotencyReplayed bool
}

type GetPaymentAddressStatusInput struct {
	Chain            valueobjects.SupportedChain
	PaymentAddressID int64
}

type GetPaymentAddressStatusResponse struct {
	PaymentAddressID        string
	AddressPolicyID         string
	ExpectedAmountMinor     int64
	Chain                   string
	Network                 string
	Scheme                  string
	MinorUnit               string
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
