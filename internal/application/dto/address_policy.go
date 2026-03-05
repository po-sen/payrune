package dto

import "payrune/internal/domain/value_objects"

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
	Chain           value_objects.Chain
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
	Chain               value_objects.Chain
	AddressPolicyID     string
	ExpectedAmountMinor int64
	CustomerReference   string
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
}
