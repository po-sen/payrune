package outbound

import (
	"context"
	"errors"
	"strings"
	"time"

	"payrune/internal/domain/valueobjects"
)

type EVMFactoryStatus string

const (
	EVMFactoryStatusActive  EVMFactoryStatus = "active"
	EVMFactoryStatusRetired EVMFactoryStatus = "retired"
)

type EVMFactoryRecord struct {
	ID                    int64
	Network               valueobjects.NetworkID
	FactoryAddress        string
	CollectorAddress      string
	VaultCreationCodeHash string
	Status                EVMFactoryStatus
	DeploymentTxHash      string
	DeployedAt            time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ReplaceActiveEVMFactoryInput struct {
	Network               valueobjects.NetworkID
	FactoryAddress        string
	CollectorAddress      string
	VaultCreationCodeHash string
	DeploymentTxHash      string
	DeployedAt            time.Time
}

func (input ReplaceActiveEVMFactoryInput) Normalize() ReplaceActiveEVMFactoryInput {
	input.Network = valueobjects.NetworkID(strings.TrimSpace(string(input.Network)))
	input.FactoryAddress = strings.TrimSpace(input.FactoryAddress)
	input.CollectorAddress = strings.TrimSpace(input.CollectorAddress)
	input.VaultCreationCodeHash = strings.TrimSpace(input.VaultCreationCodeHash)
	input.DeploymentTxHash = strings.TrimSpace(input.DeploymentTxHash)
	if !input.DeployedAt.IsZero() {
		input.DeployedAt = input.DeployedAt.UTC()
	}
	return input
}

func (input ReplaceActiveEVMFactoryInput) Validate() (ReplaceActiveEVMFactoryInput, error) {
	normalized := input.Normalize()
	if normalized.Network == "" {
		return ReplaceActiveEVMFactoryInput{}, errors.New("network is required")
	}
	if _, ok := valueobjects.ParseNetworkID(string(normalized.Network)); !ok {
		return ReplaceActiveEVMFactoryInput{}, errors.New("network is invalid")
	}
	if normalized.FactoryAddress == "" {
		return ReplaceActiveEVMFactoryInput{}, errors.New("factory address is required")
	}
	if normalized.CollectorAddress == "" {
		return ReplaceActiveEVMFactoryInput{}, errors.New("collector address is required")
	}
	if normalized.VaultCreationCodeHash == "" {
		return ReplaceActiveEVMFactoryInput{}, errors.New("vault creation code hash is required")
	}
	return normalized, nil
}

type EVMFactoryStore interface {
	ReplaceActive(
		ctx context.Context,
		input ReplaceActiveEVMFactoryInput,
		now time.Time,
	) (EVMFactoryRecord, error)
	ListActive(ctx context.Context) ([]EVMFactoryRecord, error)
	FindActiveByNetwork(
		ctx context.Context,
		network valueobjects.NetworkID,
	) (EVMFactoryRecord, bool, error)
}
