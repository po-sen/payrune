package dto

import (
	"time"

	"payrune/internal/domain/valueobjects"
)

type RegisterEVMFactoryInput struct {
	Network               valueobjects.NetworkID
	FactoryAddress        string
	CollectorAddress      string
	VaultCreationCodeHash string
	DeploymentTxHash      string
	DeployedAt            time.Time
	AllowReplaceActive    bool
}

type RegisterEVMFactoryResponse struct {
	ID                    int64     `json:"id"`
	Network               string    `json:"network"`
	FactoryAddress        string    `json:"factoryAddress"`
	CollectorAddress      string    `json:"collectorAddress"`
	VaultCreationCodeHash string    `json:"vaultCreationCodeHash"`
	Status                string    `json:"status"`
	DeploymentTxHash      string    `json:"deploymentTxHash,omitempty"`
	DeployedAt            time.Time `json:"deployedAt,omitempty"`
}
