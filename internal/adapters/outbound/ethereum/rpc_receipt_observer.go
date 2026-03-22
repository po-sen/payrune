package ethereum

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type EthereumRPCObserverConfig struct {
	Endpoint string
	Username string
	Password string
	Timeout  time.Duration
}

type EthereumRPCReceiptObserver struct {
	clients map[valueobjects.NetworkID]*ethereumRPCClient
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

type ethereumRPCBlockHeader struct {
	Number    string `json:"number"`
	Timestamp string `json:"timestamp"`
}

type ethereumRPCBlock struct {
	Number       string                   `json:"number"`
	Timestamp    string                   `json:"timestamp"`
	Transactions []ethereumRPCTransaction `json:"transactions"`
}

type ethereumRPCTransaction struct {
	Hash  string  `json:"hash"`
	To    *string `json:"to"`
	Value string  `json:"value"`
}

func NewEthereumRPCReceiptObserver(
	configs map[valueobjects.NetworkID]*EthereumRPCObserverConfig,
) (*EthereumRPCReceiptObserver, error) {
	clients := make(map[valueobjects.NetworkID]*ethereumRPCClient, len(configs))
	for rawNetwork, config := range configs {
		network, ok := valueobjects.ParseNetworkID(string(rawNetwork))
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
		return outport.ObservePaymentAddressOutput{}, err
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

	startBlockHeight, err := client.findFirstBlockOnOrAfter(ctx, input.IssuedAt.UTC(), input.LatestBlockHeight)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}
	if startBlockHeight > input.LatestBlockHeight {
		return outport.ObservePaymentAddressOutput{
			LatestBlockHeight: input.LatestBlockHeight,
		}, nil
	}

	var (
		confirmedTotalMinor   int64
		unconfirmedTotalMinor int64
	)

	for blockHeight := startBlockHeight; blockHeight <= input.LatestBlockHeight; blockHeight++ {
		block, found, err := client.fetchBlockByNumber(ctx, blockHeight, true)
		if err != nil {
			return outport.ObservePaymentAddressOutput{}, err
		}
		if !found {
			return outport.ObservePaymentAddressOutput{}, fmt.Errorf("ethereum block %d is not found", blockHeight)
		}

		blockTimestamp, err := parseEthereumHexQuantityToInt64(block.Timestamp, "block timestamp")
		if err != nil {
			return outport.ObservePaymentAddressOutput{}, err
		}
		if blockTimestamp < input.IssuedAt.UTC().Unix() {
			continue
		}

		for _, tx := range block.Transactions {
			if tx.To == nil {
				continue
			}

			toAddress, _, err := normalizeFixedHex(*tx.To, 20, "transaction to address")
			if err != nil {
				continue
			}
			if toAddress != address {
				continue
			}

			valueMinor, err := parseEthereumHexQuantityToInt64(tx.Value, "transaction value")
			if err != nil {
				return outport.ObservePaymentAddressOutput{}, err
			}
			if valueMinor <= 0 {
				continue
			}

			confirmations := calculateEthereumConfirmations(input.LatestBlockHeight, blockHeight)
			if confirmations >= int64(input.RequiredConfirmations) {
				confirmedTotalMinor, err = safeAddInt64(confirmedTotalMinor, valueMinor)
				if err != nil {
					return outport.ObservePaymentAddressOutput{}, err
				}
				continue
			}

			unconfirmedTotalMinor, err = safeAddInt64(unconfirmedTotalMinor, valueMinor)
			if err != nil {
				return outport.ObservePaymentAddressOutput{}, err
			}
		}
	}

	observedTotalMinor, err := safeAddInt64(confirmedTotalMinor, unconfirmedTotalMinor)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
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
	network valueobjects.NetworkID,
) (int64, error) {
	client, err := o.selectClient(network)
	if err != nil {
		return 0, err
	}
	return client.fetchLatestBlockHeight(ctx)
}

func (o *EthereumRPCReceiptObserver) selectClient(
	network valueobjects.NetworkID,
) (*ethereumRPCClient, error) {
	normalizedNetwork, ok := valueobjects.ParseNetworkID(string(network))
	if !ok {
		return nil, fmt.Errorf("ethereum network is invalid: %s", network)
	}

	client, ok := o.clients[normalizedNetwork]
	if !ok || client == nil {
		return nil, fmt.Errorf("ethereum %s rpc endpoint is not configured", normalizedNetwork)
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

func (c *ethereumRPCClient) findFirstBlockOnOrAfter(
	ctx context.Context,
	issuedAt time.Time,
	latestBlockHeight int64,
) (int64, error) {
	if latestBlockHeight <= 0 {
		return 0, errors.New("latest block height must be greater than zero")
	}

	latestBlock, found, err := c.fetchBlockHeaderByNumber(ctx, latestBlockHeight)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("ethereum latest block %d is not found", latestBlockHeight)
	}
	latestBlockTimestamp, err := parseEthereumHexQuantityToInt64(latestBlock.Timestamp, "block timestamp")
	if err != nil {
		return 0, err
	}
	if latestBlockTimestamp < issuedAt.Unix() {
		return latestBlockHeight + 1, nil
	}

	low := int64(0)
	high := latestBlockHeight
	for low < high {
		mid := low + (high-low)/2
		header, found, err := c.fetchBlockHeaderByNumber(ctx, mid)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, fmt.Errorf("ethereum block %d is not found", mid)
		}

		midTimestamp, err := parseEthereumHexQuantityToInt64(header.Timestamp, "block timestamp")
		if err != nil {
			return 0, err
		}
		if midTimestamp < issuedAt.Unix() {
			low = mid + 1
			continue
		}
		high = mid
	}
	return low, nil
}

func (c *ethereumRPCClient) fetchBlockHeaderByNumber(
	ctx context.Context,
	blockHeight int64,
) (ethereumRPCBlockHeader, bool, error) {
	var block *ethereumRPCBlockHeader
	if err := c.call(ctx, "eth_getBlockByNumber", []any{encodeEthereumBlockNumber(blockHeight), false}, &block); err != nil {
		return ethereumRPCBlockHeader{}, false, err
	}
	if block == nil {
		return ethereumRPCBlockHeader{}, false, nil
	}
	return *block, true, nil
}

func (c *ethereumRPCClient) fetchBlockByNumber(
	ctx context.Context,
	blockHeight int64,
	fullTransactions bool,
) (ethereumRPCBlock, bool, error) {
	var block *ethereumRPCBlock
	if err := c.call(ctx, "eth_getBlockByNumber", []any{encodeEthereumBlockNumber(blockHeight), fullTransactions}, &block); err != nil {
		return ethereumRPCBlock{}, false, err
	}
	if block == nil {
		return ethereumRPCBlock{}, false, nil
	}
	return *block, true, nil
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

func calculateEthereumConfirmations(latestBlockHeight int64, blockHeight int64) int64 {
	if latestBlockHeight <= 0 || blockHeight <= 0 || latestBlockHeight < blockHeight {
		return 0
	}
	return latestBlockHeight - blockHeight + 1
}

func safeAddInt64(left int64, right int64) (int64, error) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, errors.New("receipt total exceeds int64")
	}
	if right < 0 && left < math.MinInt64-right {
		return 0, errors.New("receipt total exceeds int64")
	}
	return left + right, nil
}

var _ outport.ChainReceiptObserver = (*EthereumRPCReceiptObserver)(nil)
