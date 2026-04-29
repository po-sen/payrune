package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	cloudflarepostgresinfra "payrune/internal/infrastructure/cloudflarepostgres"
)

const (
	cfDefaultBitcoinRequiredConfirmations = int32(2)
	cfDefaultBitcoinReceiptExpiresAfter   = 24 * time.Hour
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

func buildCloudflareAPIHTTPHandler(
	env map[string]string,
	bridgeID string,
) (http.Handler, error) {
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
	bitcoinChainAddressDeriver := bitcoin.NewChainAddressDeriver(bitcoinDeriver)
	ethereumChainAddressDeriver := ethereum.NewChainAddressDeriver()
	ethereumCreate2SaltDeriver := ethereum.NewCreate2SaltDeriver(
		buildEthereumCreate2DerivationKeys(
			cloudflareAPIEnvValue(env, envEthereumMainnetCreate2DerivationKey),
			cloudflareAPIEnvValue(env, envEthereumSepoliaCreate2DerivationKey),
		),
	)
	addressIssuancePolicies, err := buildAddressIssuancePolicies(func(key string) string {
		return cloudflareAPIEnvValue(env, key)
	}, ethereumCreate2SaltDeriver)
	if err != nil {
		return nil, err
	}
	bridge := cloudflarepostgresinfra.NewJSBridge()
	if err := validateConfiguredAddressIssuancePolicies(addressIssuancePolicies, bitcoinDeriver); err != nil {
		return nil, err
	}
	readinessChecker, err := ethereum.NewAddressIssuanceReadinessChecker(
		loadEthereumRPCConfigsFromLookup(func(key string) string {
			return cloudflareAPIEnvValue(env, key)
		}),
	)
	if err != nil {
		return nil, err
	}
	if err := validateEnabledEthereumIssuanceReadiness(context.Background(), addressIssuancePolicies, readinessChecker); err != nil {
		return nil, err
	}
	addressPolicyReader := policyadapter.NewAddressPolicyReader(addressIssuancePolicyRecords(addressIssuancePolicies))

	listAddressPoliciesUseCase := usecases.NewListAddressPoliciesUseCase(addressPolicyReader)
	dbExecutor := cloudflarepostgresadapter.NewExecutor(bridgeID, bridge)
	unitOfWork := cloudflarepostgresadapter.NewUnitOfWork(bridgeID, bridge)
	allocationIssuancePolicy := policies.NewPaymentAddressAllocationIssuancePolicy(
		requiredConfirmationsByScope,
		receiptExpiresAfterByScope,
	)
	issuedAddressDeriver, err := blockchain.NewMultiChainIssuedPaymentAddressDeriver(
		bitcoin.NewIssuedPaymentAddressDeriver(bitcoinChainAddressDeriver),
		ethereum.NewIssuedPaymentAddressDeriver(ethereumChainAddressDeriver, ethereumCreate2SaltDeriver),
	)
	if err != nil {
		return nil, err
	}
	allocatePaymentAddressUseCase := usecases.NewAllocatePaymentAddressUseCase(
		unitOfWork,
		issuedAddressDeriver,
		addressPolicyReader,
		allocationIssuancePolicy,
		clock,
	)
	getPaymentAddressStatusUseCase := usecases.NewGetPaymentAddressStatusUseCase(
		cloudflarepostgresadapter.NewPaymentAddressStatusFinder(dbExecutor),
		addressPolicyReader,
	)

	return httpadapter.NewPublicRouter(httpadapter.RouterControllers{
		Health:              httpcontroller.NewHealthController(healthUseCase),
		ListAddressPolicies: httpcontroller.NewListAddressPoliciesController(listAddressPoliciesUseCase),
		AllocatePaymentAddress: httpcontroller.NewAllocatePaymentAddressController(
			allocatePaymentAddressUseCase,
		),
		GetPaymentAddressStatus: httpcontroller.NewGetPaymentAddressStatusController(
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
	return loadReceiptRequiredConfirmationsFromLookup(
		func(key string) string { return env[key] },
		cfDefaultBitcoinRequiredConfirmations,
	)
}

func loadCloudflareReceiptExpiresAfter(
	env map[string]string,
) (map[policies.PaymentReceiptTermsScope]time.Duration, error) {
	return loadReceiptExpiresAfterByScopeFromLookup(
		func(key string) string { return env[key] },
		cfDefaultBitcoinReceiptExpiresAfter,
	)
}

func cloudflareAPIEnvValue(env map[string]string, key string) string {
	return strings.TrimSpace(env[key])
}
