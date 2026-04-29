package ethereum

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	outport "payrune/internal/application/ports/outbound"
)

var erc20BalanceOfSelector = []byte{0x70, 0xa0, 0x82, 0x31}

type EthereumRPCObserverConfig struct {
	Endpoint string
	Username string
	Password string
	Timeout  time.Duration
}

type EthereumRPCReceiptObserver struct {
	clients map[string]*ethereumRPCClient
}

type ethereumRPCClient struct {
	endpoint   string
	username   string
	password   string
	httpClient *http.Client
}

type ethereumRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type ethereumRPCResponse struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      json.RawMessage   `json:"id"`
	Result  json.RawMessage   `json:"result"`
	Error   *ethereumRPCError `json:"error"`
}

type ethereumRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewEthereumRPCReceiptObserver(
	configs map[string]*EthereumRPCObserverConfig,
) (*EthereumRPCReceiptObserver, error) {
	clients := make(map[string]*ethereumRPCClient, len(configs))
	for rawNetwork, config := range configs {
		network, ok := outport.NormalizeNetworkID(rawNetwork)
		if !ok {
			return nil, fmt.Errorf("ethereum network is invalid: %s", rawNetwork)
		}

		client, err := newEthereumRPCClient(config)
		if err != nil {
			return nil, fmt.Errorf("configure %s ethereum rpc client: %w", network, err)
		}
		if client == nil {
			continue
		}

		clients[network] = client
	}
	if len(clients) == 0 {
		return nil, errors.New("at least one ethereum rpc endpoint is required")
	}

	return &EthereumRPCReceiptObserver{clients: clients}, nil
}

func (o *EthereumRPCReceiptObserver) ObserveAddress(
	ctx context.Context,
	input outport.ObservePaymentAddressInput,
) (outport.ObservePaymentAddressOutput, error) {
	address, _, err := normalizeFixedHex(input.Address, 20, "address")
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}
	if input.IssuedAt.IsZero() {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}
	if input.RequiredConfirmations <= 0 {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}
	if input.LatestBlockHeight <= 0 {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}
	if input.SinceBlockHeight < 0 {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
	}
	assetReference := strings.TrimSpace(input.AssetReference)
	erc20AssetReference := ""
	if assetReference != "" {
		erc20AssetReference, err = NormalizeEthereumAddress(assetReference, "asset reference")
		if err != nil {
			return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverInputInvalid
		}
	}

	client, err := o.selectClient(input.Network)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}

	observedTotalMinor, err := client.fetchObservedBalanceAtBlock(
		ctx,
		assetReference,
		erc20AssetReference,
		address,
		input.LatestBlockHeight,
	)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverFailed
	}

	var confirmedTotalMinor int64
	confirmedBlockHeight := input.LatestBlockHeight - int64(input.RequiredConfirmations) + 1
	if confirmedBlockHeight > 0 {
		confirmedTotalMinor, err = client.fetchObservedBalanceAtBlock(
			ctx,
			assetReference,
			erc20AssetReference,
			address,
			confirmedBlockHeight,
		)
		if err != nil {
			return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverFailed
		}
	}

	unconfirmedTotalMinor, err := safeSubtractInt64NonNegative(observedTotalMinor, confirmedTotalMinor)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, outport.ErrBlockchainReceiptObserverFailed
	}

	return outport.ObservePaymentAddressOutput{
		ObservedTotalMinor:    observedTotalMinor,
		ConfirmedTotalMinor:   confirmedTotalMinor,
		UnconfirmedTotalMinor: unconfirmedTotalMinor,
		LatestBlockHeight:     input.LatestBlockHeight,
	}, nil
}

func (o *EthereumRPCReceiptObserver) FetchLatestBlockHeight(
	ctx context.Context,
	network string,
) (int64, error) {
	client, err := o.selectClient(network)
	if err != nil {
		return 0, err
	}
	latestBlockHeight, err := client.fetchLatestBlockHeight(ctx)
	if err != nil {
		return 0, outport.ErrBlockchainReceiptObserverFailed
	}
	return latestBlockHeight, nil
}

func (o *EthereumRPCReceiptObserver) selectClient(
	network string,
) (*ethereumRPCClient, error) {
	normalizedNetwork, ok := outport.NormalizeNetworkID(network)
	if !ok {
		return nil, outport.ErrBlockchainReceiptObserverInputInvalid
	}

	client, ok := o.clients[normalizedNetwork]
	if !ok || client == nil {
		return nil, outport.ErrBlockchainReceiptObserverNotConfigured
	}
	return client, nil
}

func newEthereumRPCClient(config *EthereumRPCObserverConfig) (*ethereumRPCClient, error) {
	if config == nil {
		return nil, nil
	}

	endpoint := strings.TrimSpace(config.Endpoint)
	if endpoint == "" {
		return nil, nil
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &ethereumRPCClient{
		endpoint: endpoint,
		username: strings.TrimSpace(config.Username),
		password: config.Password,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *ethereumRPCClient) fetchLatestBlockHeight(ctx context.Context) (int64, error) {
	var rawHeight string
	if err := c.call(ctx, "eth_blockNumber", []any{}, &rawHeight); err != nil {
		return 0, err
	}
	return parseEthereumHexQuantityToInt64(rawHeight, "latest block height")
}

func (c *ethereumRPCClient) fetchBalanceAtBlock(
	ctx context.Context,
	address string,
	blockHeight int64,
) (int64, error) {
	if blockHeight < 0 {
		return 0, errors.New("block height must be greater than or equal to zero")
	}

	var rawBalance string
	if err := c.call(ctx, "eth_getBalance", []any{address, encodeEthereumBlockNumber(blockHeight)}, &rawBalance); err != nil {
		return 0, err
	}
	return parseEthereumHexQuantityToInt64(rawBalance, "balance")
}

func (c *ethereumRPCClient) fetchObservedBalanceAtBlock(
	ctx context.Context,
	assetReference string,
	erc20AssetReference string,
	address string,
	blockHeight int64,
) (int64, error) {
	if strings.TrimSpace(assetReference) != "" {
		return c.fetchERC20BalanceAtBlock(ctx, erc20AssetReference, address, blockHeight)
	}
	return c.fetchBalanceAtBlock(ctx, address, blockHeight)
}

func (c *ethereumRPCClient) fetchERC20BalanceAtBlock(
	ctx context.Context,
	erc20AssetReference string,
	address string,
	blockHeight int64,
) (int64, error) {
	if blockHeight < 0 {
		return 0, errors.New("block height must be greater than or equal to zero")
	}

	callData, err := encodeERC20BalanceOfCall(address)
	if err != nil {
		return 0, err
	}

	var rawBalance string
	if err := c.call(
		ctx,
		"eth_call",
		[]any{
			map[string]string{
				"to":   erc20AssetReference,
				"data": callData,
			},
			encodeEthereumBlockNumber(blockHeight),
		},
		&rawBalance,
	); err != nil {
		return 0, err
	}
	return parseEthereumHexQuantityToInt64(rawBalance, "token balance")
}

func (c *ethereumRPCClient) call(
	ctx context.Context,
	method string,
	params any,
	target any,
) error {
	requestBody, err := json.Marshal(ethereumRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("ethereum rpc %s returned status %d: %s", method, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rpcResponse ethereumRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return err
	}
	if rpcResponse.Error != nil {
		return fmt.Errorf(
			"ethereum rpc %s returned error %d: %s",
			method,
			rpcResponse.Error.Code,
			strings.TrimSpace(rpcResponse.Error.Message),
		)
	}
	if len(rpcResponse.Result) == 0 {
		return errors.New("ethereum rpc response result is missing")
	}
	return json.Unmarshal(rpcResponse.Result, target)
}

func encodeEthereumBlockNumber(blockHeight int64) string {
	return "0x" + strconv.FormatInt(blockHeight, 16)
}

func parseEthereumHexQuantityToInt64(raw string, label string) (int64, error) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "0x") && !strings.HasPrefix(trimmed, "0X") {
		return 0, fmt.Errorf("%s must start with 0x", label)
	}

	value, ok := new(big.Int).SetString(trimmed[2:], 16)
	if !ok {
		return 0, fmt.Errorf("%s is invalid hex", label)
	}
	if value.Sign() < 0 {
		return 0, fmt.Errorf("%s must be non-negative", label)
	}
	if !value.IsInt64() {
		return 0, fmt.Errorf("%s exceeds int64", label)
	}
	return value.Int64(), nil
}

func safeSubtractInt64NonNegative(left int64, right int64) (int64, error) {
	if left < 0 || right < 0 {
		return 0, errors.New("receipt totals must be non-negative")
	}
	if right > left {
		return 0, errors.New("receipt totals are inconsistent")
	}
	return left - right, nil
}

func encodeERC20BalanceOfCall(address string) (string, error) {
	_, addressBytes, err := normalizeFixedHex(address, 20, "address")
	if err != nil {
		return "", err
	}

	callData := make([]byte, 4+32)
	copy(callData[:4], erc20BalanceOfSelector)
	copy(callData[4+12:], addressBytes)
	return "0x" + hex.EncodeToString(callData), nil
}

var _ outport.ChainReceiptObserver = (*EthereumRPCReceiptObserver)(nil)
