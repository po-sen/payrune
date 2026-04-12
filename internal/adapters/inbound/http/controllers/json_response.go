package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"payrune/internal/application/dto"
)

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type addressPolicyResponse struct {
	AddressPolicyID string `json:"addressPolicyId"`
	Chain           string `json:"chain"`
	Network         string `json:"network"`
	Scheme          string `json:"scheme"`
	AssetReference  string `json:"assetReference,omitempty"`
	Decimals        uint8  `json:"decimals"`
	Enabled         bool   `json:"enabled"`
}

type listAddressPoliciesResponse struct {
	Chain           string                  `json:"chain"`
	AddressPolicies []addressPolicyResponse `json:"addressPolicies"`
}

type allocatePaymentAddressResponse struct {
	PaymentAddressID    string `json:"paymentAddressId"`
	AddressPolicyID     string `json:"addressPolicyId"`
	ExpectedAmountMinor int64  `json:"expectedAmountMinor"`
	Chain               string `json:"chain"`
	Network             string `json:"network"`
	Scheme              string `json:"scheme"`
	AssetReference      string `json:"assetReference,omitempty"`
	Decimals            uint8  `json:"decimals"`
	Address             string `json:"address"`
	CustomerReference   string `json:"customerReference,omitempty"`
}

type paymentAddressStatusResponse struct {
	PaymentAddressID        string     `json:"paymentAddressId"`
	AddressPolicyID         string     `json:"addressPolicyId"`
	ExpectedAmountMinor     int64      `json:"expectedAmountMinor"`
	Chain                   string     `json:"chain"`
	Network                 string     `json:"network"`
	Scheme                  string     `json:"scheme"`
	AssetReference          string     `json:"assetReference,omitempty"`
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

func writeErrorJSON(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, errorResponse{Error: message})
}

func newHealthResponse(response dto.HealthResponse) healthResponse {
	return healthResponse{
		Status:    response.Status,
		Timestamp: response.Timestamp.UTC().Format(time.RFC3339),
	}
}

func newListAddressPoliciesResponse(response dto.ListAddressPoliciesResponse) listAddressPoliciesResponse {
	policies := make([]addressPolicyResponse, 0, len(response.AddressPolicies))
	for _, policy := range response.AddressPolicies {
		policies = append(policies, addressPolicyResponse{
			AddressPolicyID: policy.AddressPolicyID,
			Chain:           policy.Chain,
			Network:         policy.Network,
			Scheme:          policy.Scheme,
			AssetReference:  policy.AssetReference,
			Decimals:        policy.Decimals,
			Enabled:         policy.Enabled,
		})
	}

	return listAddressPoliciesResponse{
		Chain:           response.Chain,
		AddressPolicies: policies,
	}
}

func newAllocatePaymentAddressResponse(response dto.AllocatePaymentAddressResponse) allocatePaymentAddressResponse {
	return allocatePaymentAddressResponse{
		PaymentAddressID:    response.PaymentAddressID,
		AddressPolicyID:     response.AddressPolicyID,
		ExpectedAmountMinor: response.ExpectedAmountMinor,
		Chain:               response.Chain,
		Network:             response.Network,
		Scheme:              response.Scheme,
		AssetReference:      response.AssetReference,
		Decimals:            response.Decimals,
		Address:             response.Address,
		CustomerReference:   response.CustomerReference,
	}
}

func newPaymentAddressStatusResponse(response dto.GetPaymentAddressStatusResponse) paymentAddressStatusResponse {
	return paymentAddressStatusResponse{
		PaymentAddressID:        response.PaymentAddressID,
		AddressPolicyID:         response.AddressPolicyID,
		ExpectedAmountMinor:     response.ExpectedAmountMinor,
		Chain:                   response.Chain,
		Network:                 response.Network,
		Scheme:                  response.Scheme,
		AssetReference:          response.AssetReference,
		Decimals:                response.Decimals,
		Address:                 response.Address,
		CustomerReference:       response.CustomerReference,
		PaymentStatus:           response.PaymentStatus,
		ObservedTotalMinor:      response.ObservedTotalMinor,
		ConfirmedTotalMinor:     response.ConfirmedTotalMinor,
		UnconfirmedTotalMinor:   response.UnconfirmedTotalMinor,
		RequiredConfirmations:   response.RequiredConfirmations,
		LastObservedBlockHeight: response.LastObservedBlockHeight,
		IssuedAt:                response.IssuedAt,
		FirstObservedAt:         response.FirstObservedAt,
		PaidAt:                  response.PaidAt,
		ConfirmedAt:             response.ConfirmedAt,
		ExpiresAt:               response.ExpiresAt,
		LastError:               response.LastError,
	}
}
