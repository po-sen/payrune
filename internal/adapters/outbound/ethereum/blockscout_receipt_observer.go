package ethereum

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type BlockscoutObserverConfig struct {
	BaseURL string
	Timeout time.Duration
}

type BlockscoutReceiptObserver struct {
	clients map[valueobjects.NetworkID]*blockscoutClient
}

type blockscoutClient struct {
	apiURL     string
	rpcURL     string
	httpClient *http.Client
}

type blockscoutEnvelope struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

type blockscoutTx struct {
	Hash            string `json:"hash"`
	TimeStamp       string `json:"timeStamp"`
	BlockNumber     string `json:"blockNumber"`
	To              string `json:"to"`
	Value           string `json:"value"`
	Confirmations   string `json:"confirmations"`
	IsError         string `json:"isError"`
	TxReceiptStatus string `json:"txreceipt_status"`
}

type blockscoutTokenTransfer struct {
	Hash            string `json:"hash"`
	TimeStamp       string `json:"timeStamp"`
	BlockNumber     string `json:"blockNumber"`
	To              string `json:"to"`
	Value           string `json:"value"`
	Confirmations   string `json:"confirmations"`
	ContractAddress string `json:"contractAddress"`
}

type blockscoutRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type blockscoutRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      int              `json:"id"`
	Result  string           `json:"result"`
	Error   *blockscoutError `json:"error"`
}

type blockscoutError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewBlockscoutReceiptObserver(
	configs map[valueobjects.NetworkID]*BlockscoutObserverConfig,
) (*BlockscoutReceiptObserver, error) {
	clients := make(map[valueobjects.NetworkID]*blockscoutClient, len(configs))
	for rawNetwork, config := range configs {
		network, ok := valueobjects.ParseNetworkID(string(rawNetwork))
		if !ok {
			return nil, fmt.Errorf("ethereum network is not supported: %s", rawNetwork)
		}

		client, err := newBlockscoutClient(config)
		if err != nil {
			return nil, fmt.Errorf("configure %s ethereum blockscout client: %w", network, err)
		}
		if client == nil {
			continue
		}
		clients[network] = client
	}
	if len(clients) == 0 {
		return nil, errors.New("at least one ethereum blockscout endpoint is required")
	}

	return &BlockscoutReceiptObserver{clients: clients}, nil
}

func (o *BlockscoutReceiptObserver) FetchLatestBlockHeight(
	ctx context.Context,
	network valueobjects.NetworkID,
) (int64, error) {
	client, err := o.selectClient(network)
	if err != nil {
		return 0, err
	}

	payload, err := json.Marshal(blockscoutRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_blockNumber",
		Params:  []any{},
	})
	if err != nil {
		return 0, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.rpcURL, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return 0, fmt.Errorf("blockscout rpc returned status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var rpcResponse blockscoutRPCResponse
	if err := json.NewDecoder(response.Body).Decode(&rpcResponse); err != nil {
		return 0, err
	}
	if rpcResponse.Error != nil {
		return 0, fmt.Errorf("blockscout rpc error %d: %s", rpcResponse.Error.Code, rpcResponse.Error.Message)
	}
	if rpcResponse.Result == "" {
		return 0, errors.New("blockscout rpc result is empty")
	}

	latestBlockHeight, err := strconv.ParseInt(strings.TrimPrefix(rpcResponse.Result, "0x"), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("parse blockscout latest block height: %w", err)
	}
	if latestBlockHeight <= 0 {
		return 0, errors.New("latest block height must be greater than zero")
	}

	return latestBlockHeight, nil
}

func (o *BlockscoutReceiptObserver) ObserveAddress(
	ctx context.Context,
	input outport.ObservePaymentAddressInput,
) (outport.ObservePaymentAddressOutput, error) {
	address, err := normalizeEVMAddress(input.Address)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, fmt.Errorf("address is invalid: %w", err)
	}
	if input.IssuedAt.IsZero() {
		return outport.ObservePaymentAddressOutput{}, errors.New("issued at is required")
	}
	if input.RequiredConfirmations <= 0 {
		return outport.ObservePaymentAddressOutput{}, errors.New("required confirmations must be greater than zero")
	}
	if input.LatestBlockHeight <= 0 {
		return outport.ObservePaymentAddressOutput{}, errors.New("latest block height must be greater than zero")
	}
	if input.SinceBlockHeight < 0 {
		return outport.ObservePaymentAddressOutput{}, errors.New("since block height must be non-negative")
	}

	client, err := o.selectClient(input.Network)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}

	assetType := strings.ToLower(strings.TrimSpace(input.AssetType))
	switch assetType {
	case "", "native":
		transfers, err := client.fetchNativeTransactions(ctx, address, input.SinceBlockHeight)
		if err != nil {
			return outport.ObservePaymentAddressOutput{}, err
		}
		confirmed, unconfirmed, err := aggregateNativeTransfers(
			address,
			input.IssuedAt.UTC(),
			int64(input.RequiredConfirmations),
			input.LatestBlockHeight,
			transfers,
		)
		if err != nil {
			return outport.ObservePaymentAddressOutput{}, err
		}
		return outport.ObservePaymentAddressOutput{
			ObservedTotalMinor:    confirmed + unconfirmed,
			ConfirmedTotalMinor:   confirmed,
			UnconfirmedTotalMinor: unconfirmed,
			LatestBlockHeight:     input.LatestBlockHeight,
		}, nil
	case "erc20":
		tokenAddress, err := normalizeEVMAddress(input.TokenAddress)
		if err != nil {
			return outport.ObservePaymentAddressOutput{}, fmt.Errorf("token address is invalid: %w", err)
		}
		transfers, err := client.fetchTokenTransfers(ctx, address, tokenAddress, input.SinceBlockHeight)
		if err != nil {
			return outport.ObservePaymentAddressOutput{}, err
		}
		confirmed, unconfirmed, err := aggregateTokenTransfers(
			address,
			tokenAddress,
			input.IssuedAt.UTC(),
			int64(input.RequiredConfirmations),
			input.LatestBlockHeight,
			transfers,
		)
		if err != nil {
			return outport.ObservePaymentAddressOutput{}, err
		}
		return outport.ObservePaymentAddressOutput{
			ObservedTotalMinor:    confirmed + unconfirmed,
			ConfirmedTotalMinor:   confirmed,
			UnconfirmedTotalMinor: unconfirmed,
			LatestBlockHeight:     input.LatestBlockHeight,
		}, nil
	default:
		return outport.ObservePaymentAddressOutput{}, fmt.Errorf("unsupported ethereum asset type: %s", assetType)
	}
}

func (o *BlockscoutReceiptObserver) selectClient(
	network valueobjects.NetworkID,
) (*blockscoutClient, error) {
	normalizedNetwork, ok := valueobjects.ParseNetworkID(string(network))
	if !ok {
		return nil, fmt.Errorf("ethereum network is not supported: %s", network)
	}
	client, ok := o.clients[normalizedNetwork]
	if !ok || client == nil {
		return nil, fmt.Errorf("ethereum %s blockscout endpoint is not configured", normalizedNetwork)
	}
	return client, nil
}

func newBlockscoutClient(config *BlockscoutObserverConfig) (*blockscoutClient, error) {
	if config == nil {
		return nil, nil
	}

	baseURL := normalizeBlockscoutBaseURL(config.BaseURL)
	if baseURL == "" {
		return nil, nil
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &blockscoutClient{
		apiURL: baseURL + "/api",
		rpcURL: baseURL + "/api/eth-rpc",
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func normalizeBlockscoutBaseURL(raw string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return ""
	}
	return strings.TrimSuffix(trimmed, "/api")
}

func (c *blockscoutClient) fetchNativeTransactions(
	ctx context.Context,
	address string,
	sinceBlockHeight int64,
) ([]blockscoutTx, error) {
	query := url.Values{}
	query.Set("module", "account")
	query.Set("action", "txlist")
	query.Set("address", address)
	query.Set("sort", "asc")
	if sinceBlockHeight > 0 {
		query.Set("startblock", strconv.FormatInt(sinceBlockHeight, 10))
	}

	var transfers []blockscoutTx
	if err := c.getAPI(ctx, query, &transfers); err != nil {
		return nil, err
	}
	return transfers, nil
}

func (c *blockscoutClient) fetchTokenTransfers(
	ctx context.Context,
	address string,
	tokenAddress string,
	sinceBlockHeight int64,
) ([]blockscoutTokenTransfer, error) {
	query := url.Values{}
	query.Set("module", "account")
	query.Set("action", "tokentx")
	query.Set("address", address)
	query.Set("contractaddress", tokenAddress)
	query.Set("sort", "asc")
	if sinceBlockHeight > 0 {
		query.Set("startblock", strconv.FormatInt(sinceBlockHeight, 10))
	}

	var transfers []blockscoutTokenTransfer
	if err := c.getAPI(ctx, query, &transfers); err != nil {
		return nil, err
	}
	return transfers, nil
}

func (c *blockscoutClient) getAPI(ctx context.Context, query url.Values, target any) error {
	endpoint := c.apiURL + "?" + query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("blockscout api returned status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var envelope blockscoutEnvelope
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return err
	}

	if len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return nil
	}
	if envelope.Status == "0" {
		message := strings.ToLower(strings.TrimSpace(envelope.Message))
		result := strings.ToLower(strings.TrimSpace(string(envelope.Result)))
		if strings.Contains(message, "no transactions found") || result == "[]" || result == "\"\"" {
			return nil
		}
	}

	if err := json.Unmarshal(envelope.Result, target); err != nil {
		return err
	}
	return nil
}

func aggregateNativeTransfers(
	address string,
	issuedAt time.Time,
	requiredConfirmations int64,
	latestBlockHeight int64,
	transfers []blockscoutTx,
) (int64, int64, error) {
	var confirmedTotal int64
	var unconfirmedTotal int64

	for _, transfer := range transfers {
		if strings.ToLower(strings.TrimSpace(transfer.To)) != address {
			continue
		}
		if transfer.IsError == "1" || transfer.TxReceiptStatus == "0" {
			continue
		}
		timestamp, err := parseUnixString(transfer.TimeStamp)
		if err != nil {
			return 0, 0, fmt.Errorf("parse native transfer timestamp: %w", err)
		}
		if timestamp < issuedAt.Unix() {
			continue
		}
		value, err := strconv.ParseInt(strings.TrimSpace(transfer.Value), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse native transfer value: %w", err)
		}
		if value <= 0 {
			continue
		}
		confirmations, err := parseConfirmations(transfer.Confirmations, transfer.BlockNumber, latestBlockHeight)
		if err != nil {
			return 0, 0, err
		}
		if confirmations >= requiredConfirmations {
			confirmedTotal += value
			continue
		}
		unconfirmedTotal += value
	}

	return confirmedTotal, unconfirmedTotal, nil
}

func aggregateTokenTransfers(
	address string,
	tokenAddress string,
	issuedAt time.Time,
	requiredConfirmations int64,
	latestBlockHeight int64,
	transfers []blockscoutTokenTransfer,
) (int64, int64, error) {
	var confirmedTotal int64
	var unconfirmedTotal int64

	for _, transfer := range transfers {
		if strings.ToLower(strings.TrimSpace(transfer.To)) != address {
			continue
		}
		if strings.ToLower(strings.TrimSpace(transfer.ContractAddress)) != tokenAddress {
			continue
		}
		timestamp, err := parseUnixString(transfer.TimeStamp)
		if err != nil {
			return 0, 0, fmt.Errorf("parse token transfer timestamp: %w", err)
		}
		if timestamp < issuedAt.Unix() {
			continue
		}
		value, err := strconv.ParseInt(strings.TrimSpace(transfer.Value), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse token transfer value: %w", err)
		}
		if value <= 0 {
			continue
		}
		confirmations, err := parseConfirmations(transfer.Confirmations, transfer.BlockNumber, latestBlockHeight)
		if err != nil {
			return 0, 0, err
		}
		if confirmations >= requiredConfirmations {
			confirmedTotal += value
			continue
		}
		unconfirmedTotal += value
	}

	return confirmedTotal, unconfirmedTotal, nil
}

func parseUnixString(raw string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, errors.New("unix timestamp must be greater than zero")
	}
	return value, nil
}

func parseConfirmations(raw string, blockNumber string, latestBlockHeight int64) (int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed != "" {
		value, err := strconv.ParseInt(trimmed, 10, 64)
		if err == nil && value >= 0 {
			return value, nil
		}
	}

	blockHeight, err := strconv.ParseInt(strings.TrimSpace(blockNumber), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse block number: %w", err)
	}
	if blockHeight <= 0 {
		return 0, errors.New("block number must be greater than zero")
	}
	if latestBlockHeight < blockHeight {
		return 0, nil
	}
	return latestBlockHeight - blockHeight + 1, nil
}

func normalizeEVMAddress(raw string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if !strings.HasPrefix(trimmed, "0x") {
		return "", errors.New("missing 0x prefix")
	}
	if len(trimmed) != 42 {
		return "", errors.New("expected 20-byte hex address")
	}
	for i := 2; i < len(trimmed); i++ {
		char := trimmed[i]
		if char >= '0' && char <= '9' {
			continue
		}
		if char >= 'a' && char <= 'f' {
			continue
		}
		return "", errors.New("address contains non-hex characters")
	}
	return trimmed, nil
}

var _ outport.ChainReceiptObserver = (*BlockscoutReceiptObserver)(nil)
