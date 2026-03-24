package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"time"

	httpadapter "payrune/internal/adapters/inbound/http"
	httpcontroller "payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/outbound/bitcoin"
	"payrune/internal/adapters/outbound/blockchain"
	"payrune/internal/adapters/outbound/ethereum"
	cloudflarepostgresadapter "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	policyadapter "payrune/internal/adapters/outbound/policy"
	"payrune/internal/adapters/outbound/system"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
	cloudflarepostgresinfra "payrune/internal/infrastructure/cloudflarepostgres"
)

const (
	envBitcoinMainnetLegacyXPub               = "BITCOIN_MAINNET_LEGACY_XPUB"
	envBitcoinMainnetSegwitXPub               = "BITCOIN_MAINNET_SEGWIT_XPUB"
	envBitcoinMainnetNativeSegwitXPub         = "BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"
	envBitcoinMainnetTaprootXPub              = "BITCOIN_MAINNET_TAPROOT_XPUB"
	envBitcoinTestnet4LegacyXPub              = "BITCOIN_TESTNET4_LEGACY_XPUB"
	envBitcoinTestnet4SegwitXPub              = "BITCOIN_TESTNET4_SEGWIT_XPUB"
	envBitcoinTestnet4NativeSegwitXPub        = "BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"
	envBitcoinTestnet4TaprootXPub             = "BITCOIN_TESTNET4_TAPROOT_XPUB"
	cfEnvBitcoinMainnetRequiredConfirmations  = "BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS"
	cfEnvBitcoinTestnet4RequiredConfirmations = "BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS"
	cfEnvBitcoinMainnetReceiptExpiresAfter    = "BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER"
	cfEnvBitcoinTestnet4ReceiptExpiresAfter   = "BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER"
	cfEnvEthereumMainnetRequiredConfirmations = "ETHEREUM_MAINNET_REQUIRED_CONFIRMATIONS"
	cfEnvEthereumSepoliaRequiredConfirmations = "ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS"
	cfEnvEthereumMainnetReceiptExpiresAfter   = "ETHEREUM_MAINNET_RECEIPT_EXPIRES_AFTER"
	cfEnvEthereumSepoliaReceiptExpiresAfter   = "ETHEREUM_SEPOLIA_RECEIPT_EXPIRES_AFTER"
	cfEnvEthereumMainnetCreate2Collector      = "ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS"
	cfEnvEthereumSepoliaCreate2Collector      = "ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS"
	cfEnvEthereumMainnetCreate2DerivationKey  = "ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY"
	cfEnvEthereumSepoliaCreate2DerivationKey  = "ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY"
	cfDefaultBitcoinRequiredConfirmations     = int32(2)
	cfDefaultBitcoinReceiptExpiresAfter       = 24 * time.Hour
)

type apiWorkerRequestEnvelope struct {
	Request  apiWorkerRequest  `json:"request"`
	Env      map[string]string `json:"env"`
	BridgeID string            `json:"bridgeId"`
}

type apiWorkerResponseEnvelope struct {
	Response apiWorkerResponse `json:"response"`
}

type apiWorkerRequest struct {
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	RawQuery string            `json:"rawQuery"`
	Headers  map[string]string `json:"headers"`
	Body     string            `json:"body"`
}

type apiWorkerResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func HandleCloudflareAPIRequestJSON(ctx context.Context, payload string) (string, error) {
	var envelope apiWorkerRequestEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}

	response, err := handleCloudflareAPIRequest(ctx, envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(apiWorkerResponseEnvelope{Response: response})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleCloudflareAPIRequest(
	ctx context.Context,
	envelope apiWorkerRequestEnvelope,
) (apiWorkerResponse, error) {
	handler, err := buildCloudflareAPIHTTPHandler(envelope.Env, envelope.BridgeID)
	if err != nil {
		return apiWorkerResponse{}, err
	}

	return executeAPIWorkerRequest(ctx, handler, envelope.Request)
}

func buildCloudflareAPIHTTPHandler(env map[string]string, bridgeID string) (http.Handler, error) {
	clock := system.NewClock()
	healthUseCase := usecases.NewCheckHealthUseCase(clock)

	requiredConfirmationsByScope, err := loadCloudflareReceiptRequiredConfirmations(env)
	if err != nil {
		return nil, err
	}
	receiptExpiresAfterByScope, err := loadCloudflareReceiptExpiresAfter(env)
	if err != nil {
		return nil, err
	}

	bitcoinDeriver := bitcoin.NewHDXPubAddressDeriver(
		bitcoin.NewLegacyAddressEncoder(),
		bitcoin.NewSegwitAddressEncoder(),
		bitcoin.NewNativeSegwitAddressEncoder(),
		bitcoin.NewTaprootAddressEncoder(),
	)
	chainAddressDeriver, err := blockchain.NewMultiChainAddressDeriver(
		bitcoin.NewChainAddressDeriver(bitcoinDeriver),
		ethereum.NewChainAddressDeriver(),
	)
	if err != nil {
		return nil, err
	}
	ethereumCreate2SaltDeriver := ethereum.NewCreate2SaltDeriver(
		buildEthereumCreate2DerivationKeys(
			cloudflareAPIEnvValue(env, cfEnvEthereumMainnetCreate2DerivationKey),
			cloudflareAPIEnvValue(env, cfEnvEthereumSepoliaCreate2DerivationKey),
		),
	)

	addressPolicyReader := policyadapter.NewAddressPolicyReader(
		buildAddressIssuancePolicies(func(key string) string {
			return cloudflareAPIEnvValue(env, key)
		}, ethereumCreate2SaltDeriver),
	)

	listAddressPoliciesUseCase := usecases.NewListAddressPoliciesUseCase(addressPolicyReader)
	generateAddressUseCase := usecases.NewGenerateAddressUseCase(chainAddressDeriver, addressPolicyReader)
	bridge := cloudflarepostgresinfra.NewJSBridge()
	dbExecutor := cloudflarepostgresadapter.NewExecutor(bridgeID, bridge)
	unitOfWork := cloudflarepostgresadapter.NewUnitOfWork(bridgeID, bridge)
	allocationIssuancePolicy := policies.NewPaymentAddressAllocationIssuancePolicy(
		requiredConfirmationsByScope,
		receiptExpiresAfterByScope,
	)
	allocatePaymentAddressUseCase := usecases.NewAllocatePaymentAddressUseCase(
		unitOfWork,
		chainAddressDeriver,
		ethereumCreate2SaltDeriver,
		addressPolicyReader,
		allocationIssuancePolicy,
		clock,
	)
	getPaymentAddressStatusUseCase := usecases.NewGetPaymentAddressStatusUseCase(
		cloudflarepostgresadapter.NewPaymentAddressStatusFinder(dbExecutor),
		addressPolicyReader,
	)

	return httpadapter.NewPublicRouter(httpadapter.RouterControllers{
		Health: httpcontroller.NewHealthController(healthUseCase),
		ChainAddress: httpcontroller.NewChainAddressController(
			listAddressPoliciesUseCase,
			generateAddressUseCase,
			allocatePaymentAddressUseCase,
			getPaymentAddressStatusUseCase,
		),
	}), nil
}

func executeAPIWorkerRequest(
	ctx context.Context,
	handler http.Handler,
	request apiWorkerRequest,
) (apiWorkerResponse, error) {
	if handler == nil {
		return apiWorkerResponse{}, errors.New("cloudflare worker handler is not configured")
	}

	method := strings.TrimSpace(request.Method)
	if method == "" {
		method = http.MethodGet
	}
	path := strings.TrimSpace(request.Path)
	if path == "" {
		path = "/"
	}

	targetURL := &url.URL{
		Scheme:   "https",
		Host:     "worker.local",
		Path:     path,
		RawQuery: strings.TrimSpace(request.RawQuery),
	}
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		method,
		targetURL.String(),
		strings.NewReader(request.Body),
	)
	if err != nil {
		return apiWorkerResponse{}, err
	}
	for name, value := range request.Headers {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue
		}
		httpRequest.Header.Set(trimmedName, value)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httpRequest)

	result := recorder.Result()
	defer func() {
		_ = result.Body.Close()
	}()
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return apiWorkerResponse{}, err
	}

	headers := make(map[string]string, len(result.Header))
	for name, values := range result.Header {
		if len(values) == 0 {
			continue
		}
		headers[name] = strings.Join(values, ", ")
	}

	return apiWorkerResponse{
		Status:  result.StatusCode,
		Headers: headers,
		Body:    string(bodyBytes),
	}, nil
}

func loadCloudflareReceiptRequiredConfirmations(
	env map[string]string,
) (map[policies.PaymentReceiptTermsScope]int32, error) {
	mainnetConfirmations, err := parseCloudflareAPIPositiveInt32MapWithDefault(env, cfEnvBitcoinMainnetRequiredConfirmations, cfDefaultBitcoinRequiredConfirmations)
	if err != nil {
		return nil, err
	}
	testnet4Confirmations, err := parseCloudflareAPIPositiveInt32MapWithDefault(env, cfEnvBitcoinTestnet4RequiredConfirmations, cfDefaultBitcoinRequiredConfirmations)
	if err != nil {
		return nil, err
	}
	ethereumMainnetConfirmations, err := parseCloudflareAPIPositiveInt32MapWithDefault(
		env,
		cfEnvEthereumMainnetRequiredConfirmations,
		cfDefaultBitcoinRequiredConfirmations,
	)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaConfirmations, err := parseCloudflareAPIPositiveInt32MapWithDefault(
		env,
		cfEnvEthereumSepoliaRequiredConfirmations,
		cfDefaultBitcoinRequiredConfirmations,
	)
	if err != nil {
		return nil, err
	}

	return map[policies.PaymentReceiptTermsScope]int32{
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
		): mainnetConfirmations,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
		): testnet4Confirmations,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkID("mainnet"),
		): ethereumMainnetConfirmations,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkID("sepolia"),
		): ethereumSepoliaConfirmations,
	}, nil
}

func loadCloudflareReceiptExpiresAfter(
	env map[string]string,
) (map[policies.PaymentReceiptTermsScope]time.Duration, error) {
	mainnetExpiresAfter, err := parseCloudflareAPIDurationMapWithDefault(env, cfEnvBitcoinMainnetReceiptExpiresAfter, cfDefaultBitcoinReceiptExpiresAfter)
	if err != nil {
		return nil, err
	}
	testnet4ExpiresAfter, err := parseCloudflareAPIDurationMapWithDefault(env, cfEnvBitcoinTestnet4ReceiptExpiresAfter, cfDefaultBitcoinReceiptExpiresAfter)
	if err != nil {
		return nil, err
	}
	ethereumMainnetExpiresAfter, err := parseCloudflareAPIDurationMapWithDefault(
		env,
		cfEnvEthereumMainnetReceiptExpiresAfter,
		cfDefaultBitcoinReceiptExpiresAfter,
	)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaExpiresAfter, err := parseCloudflareAPIDurationMapWithDefault(
		env,
		cfEnvEthereumSepoliaReceiptExpiresAfter,
		cfDefaultBitcoinReceiptExpiresAfter,
	)
	if err != nil {
		return nil, err
	}

	return map[policies.PaymentReceiptTermsScope]time.Duration{
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
		): mainnetExpiresAfter,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
		): testnet4ExpiresAfter,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkID("mainnet"),
		): ethereumMainnetExpiresAfter,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkID("sepolia"),
		): ethereumSepoliaExpiresAfter,
	}, nil
}

func parseCloudflareAPIPositiveInt32MapWithDefault(env map[string]string, key string, fallback int32) (int32, error) {
	rawValue := cloudflareAPIEnvValue(env, key)
	if rawValue == "" {
		return fallback, nil
	}

	parsedValue, err := strconv.ParseInt(rawValue, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be a positive integer: %w", key, err)
	}
	if parsedValue <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return int32(parsedValue), nil
}

func cloudflareAPIEnvValue(env map[string]string, key string) string {
	return strings.TrimSpace(env[key])
}

func parseCloudflareAPIDurationMapWithDefault(env map[string]string, key string, fallback time.Duration) (time.Duration, error) {
	rawValue := strings.TrimSpace(env[key])
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
