package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

const defaultEVMSweepBatchSize = 50

type runEVMSweepUseCase struct {
	vaultStore outport.EVMPaymentVaultStore
	executor   outport.EVMSweepExecutor
	runtimes   map[valueobjects.NetworkID]dto.EVMSweepNetworkRuntime
}

type evmSweepGroupKey struct {
	network          valueobjects.NetworkID
	factoryAddress   string
	collectorAddress string
	assetCode        string
	assetType        string
	tokenAddress     string
}

var _ inport.RunEVMSweepUseCase = (*runEVMSweepUseCase)(nil)

func NewRunEVMSweepUseCase(
	vaultStore outport.EVMPaymentVaultStore,
	executor outport.EVMSweepExecutor,
	runtimes map[valueobjects.NetworkID]dto.EVMSweepNetworkRuntime,
) inport.RunEVMSweepUseCase {
	normalizedRuntimes := make(map[valueobjects.NetworkID]dto.EVMSweepNetworkRuntime, len(runtimes))
	for network, runtime := range runtimes {
		normalizedNetwork, ok := valueobjects.ParseNetworkID(string(network))
		if !ok {
			continue
		}
		runtime.Network = normalizedNetwork
		normalizedRuntimes[normalizedNetwork] = runtime
	}

	return &runEVMSweepUseCase{
		vaultStore: vaultStore,
		executor:   executor,
		runtimes:   normalizedRuntimes,
	}
}

func (uc *runEVMSweepUseCase) Execute(
	ctx context.Context,
	input dto.RunEVMSweepInput,
) (dto.RunEVMSweepOutput, error) {
	if uc.vaultStore == nil {
		return dto.RunEVMSweepOutput{}, errors.New("evm payment vault store is not configured")
	}
	if !input.DryRun && uc.executor == nil {
		return dto.RunEVMSweepOutput{}, errors.New("evm sweep executor is not configured")
	}

	findInput, err := outport.FindEVMSweepCandidatesInput{
		Network:           input.Network,
		AssetCode:         input.AssetCode,
		PaymentAddressIDs: input.PaymentAddressIDs,
		BeforeIssuedAt:    input.BeforeIssuedAt,
		Limit:             resolveEVMSweepBatchSize(input.BatchSize),
	}.Validate()
	if err != nil {
		return dto.RunEVMSweepOutput{}, err
	}

	candidates, err := uc.vaultStore.FindSweepCandidates(ctx, findInput)
	if err != nil {
		return dto.RunEVMSweepOutput{}, err
	}
	groups := groupEVMSweepCandidates(candidates)

	output := dto.RunEVMSweepOutput{
		CandidateCount: len(candidates),
		BatchCount:     len(groups),
		Batches:        make([]dto.RunEVMSweepBatchResult, 0, len(groups)),
	}
	if len(groups) == 0 {
		return output, nil
	}

	var executionErrors []string
	for _, group := range groups {
		batchResult, batchErr := uc.processGroup(ctx, group, input.DryRun)
		output.Batches = append(output.Batches, batchResult)
		if batchErr != nil {
			executionErrors = append(executionErrors, batchErr.Error())
		}
	}

	if len(executionErrors) > 0 {
		return output, errors.New(strings.Join(executionErrors, "; "))
	}
	return output, nil
}

func (uc *runEVMSweepUseCase) processGroup(
	ctx context.Context,
	group []outport.EVMSweepCandidateRecord,
	dryRun bool,
) (dto.RunEVMSweepBatchResult, error) {
	if len(group) == 0 {
		return dto.RunEVMSweepBatchResult{}, errors.New("evm sweep group is required")
	}

	first := group[0]
	paymentAddressIDs := make([]int64, 0, len(group))
	paymentAddressIDStrings := make([]string, 0, len(group))
	saltHexes := make([]string, 0, len(group))
	for _, candidate := range group {
		paymentAddressIDs = append(paymentAddressIDs, candidate.PaymentAddressID)
		paymentAddressIDStrings = append(paymentAddressIDStrings, strconv.FormatInt(candidate.PaymentAddressID, 10))
		saltHexes = append(saltHexes, candidate.SaltHex)
	}
	slices.Sort(paymentAddressIDStrings)

	result := dto.RunEVMSweepBatchResult{
		Network:           string(first.Network),
		FactoryAddress:    first.FactoryAddress,
		AssetCode:         first.AssetCode,
		AssetType:         first.AssetType,
		TokenAddress:      first.TokenAddress,
		PaymentAddressIDs: paymentAddressIDStrings,
		Status:            "selected",
	}
	if dryRun {
		result.Status = "dry_run"
		return result, nil
	}

	runtime, ok := uc.runtimes[first.Network]
	if !ok {
		result.Status = "failed"
		result.Error = fmt.Sprintf("ethereum %s sweeper runtime is not configured", first.Network)
		return result, errors.New(result.Error)
	}

	submission, err := uc.executor.ExecuteBatch(ctx, outport.ExecuteEVMSweepBatchInput{
		Network:           first.Network,
		RPCURL:            runtime.RPCURL,
		SweeperPrivateKey: runtime.SweeperPrivateKey,
		FactoryAddress:    first.FactoryAddress,
		AssetType:         first.AssetType,
		TokenAddress:      first.TokenAddress,
		SaltHexes:         saltHexes,
	})
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		if submission.TxHash != "" {
			result.TxHash = submission.TxHash
			if markErr := uc.vaultStore.MarkSweepFailed(ctx, outport.MarkEVMSweepResultInput{
				PaymentAddressIDs: paymentAddressIDs,
				TxHash:            submission.TxHash,
				LastError:         err.Error(),
			}); markErr != nil {
				return result, fmt.Errorf("mark evm sweep failed after submission error: %w", markErr)
			}
		}
		return result, err
	}

	result.TxHash = submission.TxHash
	if err := uc.vaultStore.MarkSweepSubmitted(ctx, outport.MarkEVMSweepSubmittedInput{
		PaymentAddressIDs: paymentAddressIDs,
		TxHash:            submission.TxHash,
	}); err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result, fmt.Errorf("mark evm sweep submitted: %w", err)
	}

	if err := uc.executor.WaitForTransaction(ctx, outport.WaitForEVMSweepTransactionInput{
		Network: first.Network,
		RPCURL:  runtime.RPCURL,
		TxHash:  submission.TxHash,
	}); err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		if markErr := uc.vaultStore.MarkSweepFailed(ctx, outport.MarkEVMSweepResultInput{
			PaymentAddressIDs: paymentAddressIDs,
			TxHash:            submission.TxHash,
			LastError:         err.Error(),
		}); markErr != nil {
			return result, fmt.Errorf("mark evm sweep failed after receipt error: %w", markErr)
		}
		return result, err
	}

	if err := uc.vaultStore.MarkSweepSucceeded(ctx, outport.MarkEVMSweepResultInput{
		PaymentAddressIDs: paymentAddressIDs,
		TxHash:            submission.TxHash,
	}); err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result, fmt.Errorf("mark evm sweep succeeded: %w", err)
	}

	result.Status = "succeeded"
	return result, nil
}

func resolveEVMSweepBatchSize(input int) int {
	if input > 0 {
		return input
	}
	return defaultEVMSweepBatchSize
}

func groupEVMSweepCandidates(candidates []outport.EVMSweepCandidateRecord) [][]outport.EVMSweepCandidateRecord {
	if len(candidates) == 0 {
		return nil
	}

	grouped := make(map[evmSweepGroupKey][]outport.EVMSweepCandidateRecord)
	order := make([]evmSweepGroupKey, 0)
	for _, candidate := range candidates {
		key := evmSweepGroupKey{
			network:          candidate.Network,
			factoryAddress:   strings.ToLower(strings.TrimSpace(candidate.FactoryAddress)),
			collectorAddress: strings.ToLower(strings.TrimSpace(candidate.CollectorAddress)),
			assetCode:        strings.ToLower(strings.TrimSpace(candidate.AssetCode)),
			assetType:        strings.ToLower(strings.TrimSpace(candidate.AssetType)),
			tokenAddress:     strings.ToLower(strings.TrimSpace(candidate.TokenAddress)),
		}
		if _, ok := grouped[key]; !ok {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], candidate)
	}

	result := make([][]outport.EVMSweepCandidateRecord, 0, len(order))
	for _, key := range order {
		result = append(result, grouped[key])
	}
	return result
}
