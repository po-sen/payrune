package bitcoin

import (
	"context"
	"errors"
	"testing"
	"time"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type fakeCloudflareEsploraBridge struct {
	latestBlockHeight int64
	chainTransactions []esploraTransaction
	mempoolTxs        []esploraTransaction
	latestErr         error
	chainErr          error
	mempoolErr        error
}

func (f *fakeCloudflareEsploraBridge) FetchLatestBlockHeight(
	context.Context,
	string,
	value_objects.NetworkID,
) (int64, error) {
	return f.latestBlockHeight, f.latestErr
}

func (f *fakeCloudflareEsploraBridge) FetchAddressChainTransactions(
	context.Context,
	string,
	value_objects.NetworkID,
	string,
) ([]esploraTransaction, error) {
	return f.chainTransactions, f.chainErr
}

func (f *fakeCloudflareEsploraBridge) FetchAddressMempoolTransactions(
	context.Context,
	string,
	value_objects.NetworkID,
	string,
) ([]esploraTransaction, error) {
	return f.mempoolTxs, f.mempoolErr
}

func TestCloudflareBitcoinEsploraReceiptObserverObserveAddress(t *testing.T) {
	issuedAt := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	observer := NewCloudflareBitcoinEsploraReceiptObserver("bridge-123", &fakeCloudflareEsploraBridge{
		chainTransactions: []esploraTransaction{
			{
				TxID: "confirmed-tx",
				Vout: []esploraTxVout{
					{Addr: "tb1qexample", Value: 1000},
				},
				Status: esploraTxStatus{
					Confirmed:   true,
					BlockHeight: 100,
					BlockTime:   issuedAt.Add(1 * time.Minute).Unix(),
				},
			},
		},
		mempoolTxs: []esploraTransaction{
			{
				TxID: "mempool-tx",
				Vout: []esploraTxVout{
					{Addr: "tb1qexample", Value: 2000},
				},
				Status: esploraTxStatus{Confirmed: false},
			},
		},
	})

	output, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
		Address:               "tb1qexample",
		IssuedAt:              issuedAt,
		RequiredConfirmations: 1,
		LatestBlockHeight:     100,
	})
	if err != nil {
		t.Fatalf("ObserveAddress returned error: %v", err)
	}
	if output.ObservedTotalMinor != 3000 {
		t.Fatalf("unexpected observed total: got %d", output.ObservedTotalMinor)
	}
	if output.ConfirmedTotalMinor != 1000 {
		t.Fatalf("unexpected confirmed total: got %d", output.ConfirmedTotalMinor)
	}
	if output.UnconfirmedTotalMinor != 2000 {
		t.Fatalf("unexpected unconfirmed total: got %d", output.UnconfirmedTotalMinor)
	}
}

func TestCloudflareBitcoinEsploraReceiptObserverFetchLatestBlockHeight(t *testing.T) {
	observer := NewCloudflareBitcoinEsploraReceiptObserver("bridge-123", &fakeCloudflareEsploraBridge{
		latestBlockHeight: 321,
	})

	latestBlockHeight, err := observer.FetchLatestBlockHeight(
		context.Background(),
		value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
	)
	if err != nil {
		t.Fatalf("FetchLatestBlockHeight returned error: %v", err)
	}
	if latestBlockHeight != 321 {
		t.Fatalf("unexpected latest block height: got %d", latestBlockHeight)
	}
}

func TestCloudflareBitcoinEsploraReceiptObserverBridgeError(t *testing.T) {
	observer := NewCloudflareBitcoinEsploraReceiptObserver("bridge-123", &fakeCloudflareEsploraBridge{
		chainErr: errors.New("boom"),
	})

	_, err := observer.ObserveAddress(context.Background(), outport.ObservePaymentAddressInput{
		Network:               value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
		Address:               "bc1qexample",
		IssuedAt:              time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC),
		RequiredConfirmations: 1,
		LatestBlockHeight:     100,
	})
	if err == nil {
		t.Fatal("expected bridge error but got nil")
	}
}
