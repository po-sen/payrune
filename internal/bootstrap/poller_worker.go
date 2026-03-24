package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	"payrune/internal/adapters/outbound/bitcoin"
	blockchainadapter "payrune/internal/adapters/outbound/blockchain"
	"payrune/internal/adapters/outbound/ethereum"
	cloudflarepostgresadapter "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	"payrune/internal/adapters/outbound/system"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
	cloudflarepostgresinfra "payrune/internal/infrastructure/cloudflarepostgres"
)

const (
	envPollRescheduleInterval = "POLL_RESCHEDULE_INTERVAL"
	envPollBatchSize          = "POLL_BATCH_SIZE"
	envPollClaimTTL           = "POLL_CLAIM_TTL"
	envPollChain              = "POLL_CHAIN"
	envPollNetwork            = "POLL_NETWORK"

	cloudflarePollerDefaultBatchSize          = 2
	cloudflarePollerDefaultRescheduleInterval = 10 * time.Minute
	cloudflarePollerDefaultClaimTTL           = 30 * time.Second
)

type pollerWorkerRequestEnvelope struct {
	Env              map[string]string `json:"env"`
	PostgresBridgeID string            `json:"postgresBridgeId"`
	BitcoinBridgeID  string            `json:"bitcoinBridgeId"`
	ScheduledTime    string            `json:"scheduledTime"`
	Cron             string            `json:"cron"`
}

type pollerWorkerResponseEnvelope struct {
	Output scheduleradapter.PollerResponse `json:"output"`
}

func HandleCloudflarePollerRequestJSON(ctx context.Context, payload string) (string, error) {
	var envelope pollerWorkerRequestEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}

	output, err := handleCloudflarePollerRequest(ctx, envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(pollerWorkerResponseEnvelope{Output: output})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleCloudflarePollerRequest(
	ctx context.Context,
	envelope pollerWorkerRequestEnvelope,
) (scheduleradapter.PollerResponse, error) {
	handler, request, err := buildCloudflarePollerRuntime(
		envelope.Env,
		envelope.PostgresBridgeID,
		envelope.BitcoinBridgeID,
	)
	if err != nil {
		return scheduleradapter.PollerResponse{}, err
	}

	return handler.Handle(ctx, request)
}

func buildCloudflarePollerRuntime(
	env map[string]string,
	postgresBridgeID string,
	bitcoinBridgeID string,
) (*scheduleradapter.PollerHandler, scheduleradapter.PollerRequest, error) {
	request, err := buildCloudflarePollerRequest(env)
	if err != nil {
		return nil, scheduleradapter.PollerRequest{}, err
	}

	clock := system.NewClock()
	unitOfWork := cloudflarepostgresadapter.NewUnitOfWork(postgresBridgeID, cloudflarepostgresinfra.NewJSBridge())
	chainObservers := make(map[valueobjects.ChainID]outport.ChainReceiptObserver, 2)
	if request.Chain == "" || request.Chain == string(valueobjects.ChainIDBitcoin) {
		chainObservers[valueobjects.ChainIDBitcoin] = bitcoin.NewCloudflareBitcoinEsploraReceiptObserver(
			bitcoinBridgeID,
			bitcoin.NewCloudflareEsploraBridge(),
		)
	}
	if request.Chain == "" || request.Chain == string(valueobjects.ChainIDEthereum) {
		if ethereumConfigs := loadEthereumRPCConfigsFromLookup(func(key string) string {
			return cloudflarePollerEnvValue(env, key)
		}); len(ethereumConfigs) > 0 {
			ethereumObserver, err := ethereum.NewEthereumRPCReceiptObserver(ethereumConfigs)
			if err != nil {
				return nil, scheduleradapter.PollerRequest{}, err
			}
			chainObservers[valueobjects.ChainIDEthereum] = ethereumObserver
		}
	}
	receiptObserver, err := blockchainadapter.NewMultiChainReceiptObserver(chainObservers)
	if err != nil {
		return nil, scheduleradapter.PollerRequest{}, err
	}

	useCase := usecases.NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		receiptObserver,
		clock,
		policies.NewPaymentReceiptTrackingLifecyclePolicy(),
	)

	handler := scheduleradapter.NewPollerHandler(scheduleradapter.PollerDependencies{
		RunReceiptPollingCycleUseCase: useCase,
	})
	return handler, request, nil
}

func buildCloudflarePollerRequest(env map[string]string) (scheduleradapter.PollerRequest, error) {
	batchSize, err := parseCloudflarePollerPositiveIntEnvWithDefault(
		env,
		envPollBatchSize,
		cloudflarePollerDefaultBatchSize,
	)
	if err != nil {
		return scheduleradapter.PollerRequest{}, err
	}
	rescheduleInterval, err := parseCloudflarePollerDurationMapWithDefault(
		env,
		envPollRescheduleInterval,
		cloudflarePollerDefaultRescheduleInterval,
	)
	if err != nil {
		return scheduleradapter.PollerRequest{}, err
	}
	claimTTL, err := parseCloudflarePollerDurationMapWithDefault(
		env,
		envPollClaimTTL,
		cloudflarePollerDefaultClaimTTL,
	)
	if err != nil {
		return scheduleradapter.PollerRequest{}, err
	}
	chain, err := parseCloudflarePollerChainEnv(env, envPollChain)
	if err != nil {
		return scheduleradapter.PollerRequest{}, err
	}
	network, err := parseCloudflarePollerNetworkEnv(env, envPollNetwork)
	if err != nil {
		return scheduleradapter.PollerRequest{}, err
	}
	if network != "" && chain == "" {
		return scheduleradapter.PollerRequest{}, fmt.Errorf("%s is required when %s is set", envPollChain, envPollNetwork)
	}

	return scheduleradapter.PollerRequest{
		BatchSize:          batchSize,
		RescheduleInterval: rescheduleInterval,
		ClaimTTL:           claimTTL,
		Chain:              chain,
		Network:            network,
	}, nil
}

func parseCloudflarePollerPositiveIntEnvWithDefault(
	env map[string]string,
	key string,
	fallback int,
) (int, error) {
	rawValue := cloudflarePollerEnvValue(env, key)
	if rawValue == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func parseCloudflarePollerChainEnv(env map[string]string, key string) (string, error) {
	rawValue := cloudflarePollerEnvValue(env, key)
	if rawValue == "" {
		return "", nil
	}

	chain, ok := valueobjects.ParseChainID(rawValue)
	if !ok {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return string(chain), nil
}

func parseCloudflarePollerNetworkEnv(env map[string]string, key string) (string, error) {
	rawValue := cloudflarePollerEnvValue(env, key)
	if rawValue == "" {
		return "", nil
	}

	network, ok := valueobjects.ParseNetworkID(rawValue)
	if !ok {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return string(network), nil
}

func cloudflarePollerEnvValue(env map[string]string, key string) string {
	return strings.TrimSpace(env[key])
}

func parseCloudflarePollerDurationMapWithDefault(
	env map[string]string,
	key string,
	fallback time.Duration,
) (time.Duration, error) {
	rawValue := cloudflarePollerEnvValue(env, key)
	if rawValue == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return duration, nil
}
