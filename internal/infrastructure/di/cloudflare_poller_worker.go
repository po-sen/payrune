package di

import (
	"fmt"
	"strconv"
	"time"

	inboundadapter "payrune/internal/adapters/inbound/cloudflareworker"
	"payrune/internal/adapters/outbound/bitcoin"
	blockchainadapter "payrune/internal/adapters/outbound/blockchain"
	cloudflarepostgres "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	"payrune/internal/adapters/outbound/system"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/application/use_cases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/value_objects"
)

const (
	envPollRescheduleInterval = "POLL_RESCHEDULE_INTERVAL"
	envPollBatchSize          = "POLL_BATCH_SIZE"
	envPollClaimTTL           = "POLL_CLAIM_TTL"
	envPollChain              = "POLL_CHAIN"
	envPollNetwork            = "POLL_NETWORK"

	defaultPollerBatchSize          = 2
	defaultPollerRescheduleInterval = 10 * time.Minute
	defaultPollerClaimTTL           = 30 * time.Second
)

func BuildCloudflarePollerRuntime(
	env map[string]string,
	postgresBridgeID string,
	bitcoinBridgeID string,
) (*inboundadapter.PollerHandler, inboundadapter.PollerRequest, error) {
	request, err := buildCloudflarePollerRequest(env)
	if err != nil {
		return nil, inboundadapter.PollerRequest{}, err
	}

	clock := system.NewClock()
	unitOfWork := cloudflarepostgres.NewUnitOfWork(postgresBridgeID, cloudflarepostgres.NewJSBridge())
	bitcoinObserver := bitcoin.NewCloudflareBitcoinEsploraReceiptObserver(
		bitcoinBridgeID,
		bitcoin.NewCloudflareEsploraBridge(),
	)
	receiptObserver, err := blockchainadapter.NewMultiChainReceiptObserver(
		map[value_objects.ChainID]outport.ChainReceiptObserver{
			value_objects.ChainIDBitcoin: bitcoinObserver,
		},
	)
	if err != nil {
		return nil, inboundadapter.PollerRequest{}, err
	}

	useCase := use_cases.NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		receiptObserver,
		clock,
		policies.NewPaymentReceiptTrackingLifecyclePolicy(),
	)

	handler := inboundadapter.NewPollerHandler(inboundadapter.PollerDependencies{
		RunReceiptPollingCycleUseCase: useCase,
	})
	return handler, request, nil
}

func buildCloudflarePollerRequest(env map[string]string) (inboundadapter.PollerRequest, error) {
	batchSize, err := parsePositiveIntEnvWithDefault(env, envPollBatchSize, defaultPollerBatchSize)
	if err != nil {
		return inboundadapter.PollerRequest{}, err
	}
	rescheduleInterval, err := parseDurationMapWithDefault(env, envPollRescheduleInterval, defaultPollerRescheduleInterval)
	if err != nil {
		return inboundadapter.PollerRequest{}, err
	}
	claimTTL, err := parseDurationMapWithDefault(env, envPollClaimTTL, defaultPollerClaimTTL)
	if err != nil {
		return inboundadapter.PollerRequest{}, err
	}
	chain, err := parseChainEnv(env, envPollChain)
	if err != nil {
		return inboundadapter.PollerRequest{}, err
	}
	network, err := parseNetworkEnv(env, envPollNetwork)
	if err != nil {
		return inboundadapter.PollerRequest{}, err
	}
	if network != "" && chain == "" {
		return inboundadapter.PollerRequest{}, fmt.Errorf("%s is required when %s is set", envPollChain, envPollNetwork)
	}

	return inboundadapter.PollerRequest{
		BatchSize:          batchSize,
		RescheduleInterval: rescheduleInterval,
		ClaimTTL:           claimTTL,
		Chain:              chain,
		Network:            network,
	}, nil
}

func parsePositiveIntEnvWithDefault(env map[string]string, key string, fallback int) (int, error) {
	rawValue := envMapValue(env, key)
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

func parseChainEnv(env map[string]string, key string) (string, error) {
	rawValue := envMapValue(env, key)
	if rawValue == "" {
		return "", nil
	}

	chain, ok := value_objects.ParseChainID(rawValue)
	if !ok {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return string(chain), nil
}

func parseNetworkEnv(env map[string]string, key string) (string, error) {
	rawValue := envMapValue(env, key)
	if rawValue == "" {
		return "", nil
	}

	network, ok := value_objects.ParseNetworkID(rawValue)
	if !ok {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return string(network), nil
}
