package dto

import (
	"time"

	"payrune/internal/domain/valueobjects"
)

type EVMSweepNetworkRuntime struct {
	Network           valueobjects.NetworkID
	RPCURL            string
	SweeperPrivateKey string
}

type RunEVMSweepInput struct {
	Network           valueobjects.NetworkID
	AssetCode         string
	PaymentAddressIDs []int64
	BeforeIssuedAt    time.Time
	BatchSize         int
	DryRun            bool
}

type RunEVMSweepBatchResult struct {
	Network           string   `json:"network"`
	FactoryAddress    string   `json:"factoryAddress"`
	AssetCode         string   `json:"assetCode"`
	AssetType         string   `json:"assetType"`
	TokenAddress      string   `json:"tokenAddress,omitempty"`
	PaymentAddressIDs []string `json:"paymentAddressIds"`
	TxHash            string   `json:"txHash,omitempty"`
	Status            string   `json:"status"`
	Error             string   `json:"error,omitempty"`
}

type RunEVMSweepOutput struct {
	CandidateCount int                      `json:"candidateCount"`
	BatchCount     int                      `json:"batchCount"`
	Batches        []RunEVMSweepBatchResult `json:"batches"`
}
