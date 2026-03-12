package dto

import (
	"time"

	"payrune/internal/domain/valueobjects"
)

type AddressPolicy struct {
	AddressPolicyID string `json:"addressPolicyId"`
	Chain           string `json:"chain"`
	Network         string `json:"network"`
	Scheme          string `json:"scheme"`
	MinorUnit       string `json:"minorUnit"`
	Decimals        uint8  `json:"decimals"`
	Enabled         bool   `json:"enabled"`
}

type ListAddressPoliciesResponse struct {
	Chain           string          `json:"chain"`
	AddressPolicies []AddressPolicy `json:"addressPolicies"`
}

type GenerateAddressInput struct {
	Chain           valueobjects.SupportedChain
	AddressPolicyID string
	Index           uint32
}

type GenerateAddressResponse struct {
	AddressPolicyID string `json:"addressPolicyId"`
	Chain           string `json:"chain"`
	Network         string `json:"network"`
	Scheme          string `json:"scheme"`
	MinorUnit       string `json:"minorUnit"`
	Decimals        uint8  `json:"decimals"`
	Index           uint32 `json:"index"`
	Address         string `json:"address"`
}

type AllocatePaymentAddressInput struct {
	Chain               valueobjects.SupportedChain
	AddressPolicyID     string
	ExpectedAmountMinor int64
	CustomerReference   string
	IdempotencyKey      string
}

type AllocatePaymentAddressResponse struct {
	PaymentAddressID    string `json:"paymentAddressId"`
	AddressPolicyID     string `json:"addressPolicyId"`
	ExpectedAmountMinor int64  `json:"expectedAmountMinor"`
	Chain               string `json:"chain"`
	Network             string `json:"network"`
	Scheme              string `json:"scheme"`
	MinorUnit           string `json:"minorUnit"`
	Decimals            uint8  `json:"decimals"`
	Address             string `json:"address"`
	CustomerReference   string `json:"customerReference,omitempty"`
	IdempotencyReplayed bool   `json:"-"`
}

type GetPaymentAddressStatusInput struct {
	Chain            valueobjects.SupportedChain
	PaymentAddressID int64
}

type GetPaymentAddressStatusResponse struct {
	PaymentAddressID        string     `json:"paymentAddressId"`
	AddressPolicyID         string     `json:"addressPolicyId"`
	ExpectedAmountMinor     int64      `json:"expectedAmountMinor"`
	Chain                   string     `json:"chain"`
	Network                 string     `json:"network"`
	Scheme                  string     `json:"scheme"`
	MinorUnit               string     `json:"minorUnit"`
	Decimals                uint8      `json:"decimals"`
	Address                 string     `json:"address"`
	CustomerReference       string     `json:"customerReference,omitempty"`
	PaymentStatus           string     `json:"paymentStatus"`
	ObservedTotalMinor      int64      `json:"observedTotalMinor"`
	ConfirmedTotalMinor     int64      `json:"confirmedTotalMinor"`
	UnconfirmedTotalMinor   int64      `json:"unconfirmedTotalMinor"`
	RequiredConfirmations   int32      `json:"requiredConfirmations"`
	LastObservedBlockHeight int64      `json:"lastObservedBlockHeight"`
	IssuedAt                time.Time  `json:"issuedAt"`
	FirstObservedAt         *time.Time `json:"firstObservedAt,omitempty"`
	PaidAt                  *time.Time `json:"paidAt,omitempty"`
	ConfirmedAt             *time.Time `json:"confirmedAt,omitempty"`
	ExpiresAt               *time.Time `json:"expiresAt,omitempty"`
	LastError               string     `json:"lastError,omitempty"`
}
