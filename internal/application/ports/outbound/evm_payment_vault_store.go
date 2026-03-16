package outbound

import (
	"context"
	"errors"
	"strings"
	"time"

	"payrune/internal/domain/valueobjects"
)

type EVMPaymentVaultRecord struct {
	PaymentAddressID int64
	Network          valueobjects.NetworkID
	FactoryID        int64
	FactoryAddress   string
	CollectorAddress string
	TokenAddress     string
	SaltHex          string
	PredictedAddress string
	DeployStatus     string
	SweepStatus      string
	DeployTxHash     string
	LastSweepTxHash  string
	LastSweepError   string
}

type EVMSweepCandidateRecord struct {
	PaymentAddressID int64
	Network          valueobjects.NetworkID
	FactoryID        int64
	FactoryAddress   string
	CollectorAddress string
	AssetCode        string
	AssetType        string
	TokenAddress     string
	SaltHex          string
	PredictedAddress string
	DeployStatus     string
	SweepStatus      string
	IssuedAt         time.Time
}

type CreateEVMPaymentVaultInput struct {
	PaymentAddressID int64
	Network          valueobjects.NetworkID
	FactoryID        int64
	FactoryAddress   string
	CollectorAddress string
	TokenAddress     string
	SaltHex          string
	PredictedAddress string
}

func (input CreateEVMPaymentVaultInput) Normalize() CreateEVMPaymentVaultInput {
	input.Network = valueobjects.NetworkID(strings.TrimSpace(string(input.Network)))
	input.FactoryAddress = strings.TrimSpace(input.FactoryAddress)
	input.CollectorAddress = strings.TrimSpace(input.CollectorAddress)
	input.TokenAddress = strings.TrimSpace(input.TokenAddress)
	input.SaltHex = strings.TrimSpace(strings.ToLower(input.SaltHex))
	input.PredictedAddress = strings.TrimSpace(input.PredictedAddress)
	return input
}

func (input CreateEVMPaymentVaultInput) Validate() (CreateEVMPaymentVaultInput, error) {
	normalized := input.Normalize()
	if normalized.PaymentAddressID <= 0 {
		return CreateEVMPaymentVaultInput{}, errors.New("payment address id is required")
	}
	if normalized.Network == "" {
		return CreateEVMPaymentVaultInput{}, errors.New("network is required")
	}
	if normalized.FactoryID <= 0 {
		return CreateEVMPaymentVaultInput{}, errors.New("factory id is required")
	}
	if normalized.FactoryAddress == "" {
		return CreateEVMPaymentVaultInput{}, errors.New("factory address is required")
	}
	if normalized.CollectorAddress == "" {
		return CreateEVMPaymentVaultInput{}, errors.New("collector address is required")
	}
	if normalized.SaltHex == "" {
		return CreateEVMPaymentVaultInput{}, errors.New("salt hex is required")
	}
	if normalized.PredictedAddress == "" {
		return CreateEVMPaymentVaultInput{}, errors.New("predicted address is required")
	}
	return normalized, nil
}

type FindEVMSweepCandidatesInput struct {
	Network           valueobjects.NetworkID
	AssetCode         string
	PaymentAddressIDs []int64
	BeforeIssuedAt    time.Time
	Limit             int
}

func (input FindEVMSweepCandidatesInput) Normalize() FindEVMSweepCandidatesInput {
	input.Network = valueobjects.NetworkID(strings.TrimSpace(string(input.Network)))
	input.AssetCode = strings.ToLower(strings.TrimSpace(input.AssetCode))
	if !input.BeforeIssuedAt.IsZero() {
		input.BeforeIssuedAt = input.BeforeIssuedAt.UTC()
	}
	normalizedIDs := make([]int64, 0, len(input.PaymentAddressIDs))
	for _, paymentAddressID := range input.PaymentAddressIDs {
		if paymentAddressID <= 0 {
			continue
		}
		normalizedIDs = append(normalizedIDs, paymentAddressID)
	}
	input.PaymentAddressIDs = normalizedIDs
	return input
}

func (input FindEVMSweepCandidatesInput) Validate() (FindEVMSweepCandidatesInput, error) {
	normalized := input.Normalize()
	if normalized.Network != "" {
		if _, ok := valueobjects.ParseNetworkID(string(normalized.Network)); !ok {
			return FindEVMSweepCandidatesInput{}, errors.New("network is invalid")
		}
	}
	if normalized.Limit < 0 {
		return FindEVMSweepCandidatesInput{}, errors.New("limit must be greater than or equal to zero")
	}
	return normalized, nil
}

type MarkEVMSweepSubmittedInput struct {
	PaymentAddressIDs []int64
	TxHash            string
}

type MarkEVMSweepResultInput struct {
	PaymentAddressIDs []int64
	TxHash            string
	LastError         string
}

func (input MarkEVMSweepSubmittedInput) Validate() (MarkEVMSweepSubmittedInput, error) {
	normalized := MarkEVMSweepSubmittedInput{
		PaymentAddressIDs: normalizePaymentAddressIDs(input.PaymentAddressIDs),
		TxHash:            strings.TrimSpace(input.TxHash),
	}
	if len(normalized.PaymentAddressIDs) == 0 {
		return MarkEVMSweepSubmittedInput{}, errors.New("payment address ids are required")
	}
	if normalized.TxHash == "" {
		return MarkEVMSweepSubmittedInput{}, errors.New("tx hash is required")
	}
	return normalized, nil
}

func (input MarkEVMSweepResultInput) Validate() (MarkEVMSweepResultInput, error) {
	normalized := MarkEVMSweepResultInput{
		PaymentAddressIDs: normalizePaymentAddressIDs(input.PaymentAddressIDs),
		TxHash:            strings.TrimSpace(input.TxHash),
		LastError:         strings.TrimSpace(input.LastError),
	}
	if len(normalized.PaymentAddressIDs) == 0 {
		return MarkEVMSweepResultInput{}, errors.New("payment address ids are required")
	}
	return normalized, nil
}

func normalizePaymentAddressIDs(ids []int64) []int64 {
	normalized := make([]int64, 0, len(ids))
	for _, paymentAddressID := range ids {
		if paymentAddressID <= 0 {
			continue
		}
		normalized = append(normalized, paymentAddressID)
	}
	return normalized
}

type EVMPaymentVaultStore interface {
	Create(ctx context.Context, input CreateEVMPaymentVaultInput) (EVMPaymentVaultRecord, error)
	FindSweepCandidates(ctx context.Context, input FindEVMSweepCandidatesInput) ([]EVMSweepCandidateRecord, error)
	MarkSweepSubmitted(ctx context.Context, input MarkEVMSweepSubmittedInput) error
	MarkSweepSucceeded(ctx context.Context, input MarkEVMSweepResultInput) error
	MarkSweepFailed(ctx context.Context, input MarkEVMSweepResultInput) error
}
