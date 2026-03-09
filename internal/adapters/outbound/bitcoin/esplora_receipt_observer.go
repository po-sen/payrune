package bitcoin

import (
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

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

const defaultEsploraChainPageSize = 25

type BitcoinEsploraObserverConfig struct {
	Endpoint string
	Username string
	Password string
	Timeout  time.Duration
}

type BitcoinEsploraReceiptObserver struct {
	clients map[value_objects.NetworkID]*bitcoinAPIClient
}

type bitcoinAPIClient struct {
	endpoint   string
	username   string
	password   string
	httpClient *http.Client
}

type esploraTransaction struct {
	TxID   string          `json:"txid"`
	Vout   []esploraTxVout `json:"vout"`
	Status esploraTxStatus `json:"status"`
}

type esploraTxVout struct {
	Value int64  `json:"value"`
	Addr  string `json:"scriptpubkey_address"`
}

type esploraTxStatus struct {
	Confirmed   bool  `json:"confirmed"`
	BlockHeight int64 `json:"block_height"`
	BlockTime   int64 `json:"block_time"`
}

func NewBitcoinEsploraReceiptObserver(
	configs map[value_objects.NetworkID]*BitcoinEsploraObserverConfig,
) (*BitcoinEsploraReceiptObserver, error) {
	clients := make(map[value_objects.NetworkID]*bitcoinAPIClient, len(configs))
	for rawNetwork, config := range configs {
		bitcoinNetwork, ok := value_objects.ParseBitcoinNetwork(string(rawNetwork))
		if !ok {
			return nil, fmt.Errorf("bitcoin network is not supported: %s", rawNetwork)
		}

		client, err := newBitcoinAPIClient(config)
		if err != nil {
			return nil, fmt.Errorf("configure %s bitcoin endpoint client: %w", bitcoinNetwork, err)
		}
		if client == nil {
			continue
		}

		clients[value_objects.NetworkID(bitcoinNetwork)] = client
	}
	if len(clients) == 0 {
		return nil, errors.New("at least one bitcoin endpoint is required")
	}

	return &BitcoinEsploraReceiptObserver{
		clients: clients,
	}, nil
}

func (o *BitcoinEsploraReceiptObserver) ObserveAddress(
	ctx context.Context,
	input outport.ObservePaymentAddressInput,
) (outport.ObservePaymentAddressOutput, error) {
	address := strings.TrimSpace(input.Address)
	if address == "" {
		return outport.ObservePaymentAddressOutput{}, errors.New("address is required")
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

	chainTransactions, err := client.fetchAddressChainTransactions(ctx, address)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}
	mempoolTransactions, err := client.fetchAddressMempoolTransactions(ctx, address)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}

	confirmedTotalMinor, unconfirmedTotalMinor, err := aggregateInboundTotals(
		address,
		input.IssuedAt.UTC(),
		int64(input.RequiredConfirmations),
		input.LatestBlockHeight,
		chainTransactions,
		mempoolTransactions,
	)
	if err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}

	return outport.ObservePaymentAddressOutput{
		ObservedTotalMinor:    confirmedTotalMinor + unconfirmedTotalMinor,
		ConfirmedTotalMinor:   confirmedTotalMinor,
		UnconfirmedTotalMinor: unconfirmedTotalMinor,
		LatestBlockHeight:     input.LatestBlockHeight,
	}, nil
}

func (o *BitcoinEsploraReceiptObserver) FetchLatestBlockHeight(
	ctx context.Context,
	network value_objects.NetworkID,
) (int64, error) {
	client, err := o.selectClient(network)
	if err != nil {
		return 0, err
	}
	return client.fetchLatestBlockHeight(ctx)
}

func (o *BitcoinEsploraReceiptObserver) selectClient(
	network value_objects.NetworkID,
) (*bitcoinAPIClient, error) {
	bitcoinNetwork, ok := value_objects.ParseBitcoinNetwork(string(network))
	if !ok {
		return nil, fmt.Errorf("bitcoin network is not supported: %s", network)
	}

	client, ok := o.clients[value_objects.NetworkID(bitcoinNetwork)]
	if !ok || client == nil {
		return nil, fmt.Errorf("bitcoin %s endpoint is not configured", bitcoinNetwork)
	}
	return client, nil
}

func newBitcoinAPIClient(config *BitcoinEsploraObserverConfig) (*bitcoinAPIClient, error) {
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

	return &bitcoinAPIClient{
		endpoint: endpoint,
		username: strings.TrimSpace(config.Username),
		password: config.Password,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func aggregateInboundTotals(
	address string,
	issuedAt time.Time,
	requiredConfirmations int64,
	latestBlockHeight int64,
	chainTransactions []esploraTransaction,
	mempoolTransactions []esploraTransaction,
) (int64, int64, error) {
	address = strings.TrimSpace(address)
	issuedAtUnix := issuedAt.Unix()

	var (
		confirmedTotalMinor   int64
		unconfirmedTotalMinor int64
	)

	seenTxIDs := make(map[string]struct{}, len(chainTransactions)+len(mempoolTransactions))
	processTransaction := func(tx esploraTransaction) error {
		txID := strings.TrimSpace(tx.TxID)
		if txID != "" {
			if _, exists := seenTxIDs[txID]; exists {
				return nil
			}
			seenTxIDs[txID] = struct{}{}
		}

		inboundTotalMinor := inboundValueForAddress(tx, address)
		if inboundTotalMinor == 0 {
			return nil
		}

		if !tx.Status.Confirmed {
			unconfirmedTotalMinor += inboundTotalMinor
			return nil
		}

		if tx.Status.BlockTime <= 0 {
			return errors.New("confirmed transaction is missing block time")
		}
		if tx.Status.BlockTime < issuedAtUnix {
			return nil
		}

		confirmations := calculateConfirmations(latestBlockHeight, tx.Status.BlockHeight)
		if confirmations >= requiredConfirmations {
			confirmedTotalMinor += inboundTotalMinor
			return nil
		}
		unconfirmedTotalMinor += inboundTotalMinor
		return nil
	}

	for _, tx := range chainTransactions {
		if err := processTransaction(tx); err != nil {
			return 0, 0, err
		}
	}
	for _, tx := range mempoolTransactions {
		if err := processTransaction(tx); err != nil {
			return 0, 0, err
		}
	}

	return confirmedTotalMinor, unconfirmedTotalMinor, nil
}

func inboundValueForAddress(tx esploraTransaction, address string) int64 {
	var total int64
	for _, output := range tx.Vout {
		if strings.TrimSpace(output.Addr) != address {
			continue
		}
		if output.Value <= 0 {
			continue
		}
		total += output.Value
	}
	return total
}

func calculateConfirmations(latestBlockHeight int64, blockHeight int64) int64 {
	if latestBlockHeight <= 0 || blockHeight <= 0 || blockHeight > latestBlockHeight {
		return 0
	}
	return latestBlockHeight - blockHeight + 1
}

func (c *bitcoinAPIClient) fetchLatestBlockHeight(ctx context.Context) (int64, error) {
	body, err := c.get(ctx, "/blocks/tip/height")
	if err != nil {
		return 0, err
	}

	heightRaw := strings.TrimSpace(string(body))
	height, err := strconv.ParseInt(heightRaw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse latest block height: %w", err)
	}
	if height < 0 {
		return 0, errors.New("latest block height must be non-negative")
	}
	return height, nil
}

func (c *bitcoinAPIClient) fetchAddressChainTransactions(
	ctx context.Context,
	address string,
) ([]esploraTransaction, error) {
	encodedAddress := url.PathEscape(strings.TrimSpace(address))
	path := fmt.Sprintf("/address/%s/txs/chain", encodedAddress)

	transactions := make([]esploraTransaction, 0)
	lastSeenTxID := ""
	for page := 0; page < 10_000; page++ {
		pagePath := path
		if lastSeenTxID != "" {
			pagePath = fmt.Sprintf("%s/%s", path, url.PathEscape(lastSeenTxID))
		}

		var pageTransactions []esploraTransaction
		if err := c.getJSON(ctx, pagePath, &pageTransactions); err != nil {
			return nil, err
		}
		if len(pageTransactions) == 0 {
			break
		}

		transactions = append(transactions, pageTransactions...)
		if len(pageTransactions) < defaultEsploraChainPageSize {
			break
		}

		nextTxID := strings.TrimSpace(pageTransactions[len(pageTransactions)-1].TxID)
		if nextTxID == "" || nextTxID == lastSeenTxID {
			break
		}
		lastSeenTxID = nextTxID
	}

	return transactions, nil
}

func (c *bitcoinAPIClient) fetchAddressMempoolTransactions(
	ctx context.Context,
	address string,
) ([]esploraTransaction, error) {
	encodedAddress := url.PathEscape(strings.TrimSpace(address))
	path := fmt.Sprintf("/address/%s/txs/mempool", encodedAddress)

	var transactions []esploraTransaction
	if err := c.getJSON(ctx, path, &transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (c *bitcoinAPIClient) getJSON(
	ctx context.Context,
	path string,
	result any,
) error {
	body, err := c.get(ctx, path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("decode endpoint response: %w", err)
	}
	return nil
}

func (c *bitcoinAPIClient) get(ctx context.Context, path string) ([]byte, error) {
	endpoint := strings.TrimRight(c.endpoint, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bitcoin endpoint call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bitcoin endpoint http status %d", resp.StatusCode)
	}

	return body, nil
}
