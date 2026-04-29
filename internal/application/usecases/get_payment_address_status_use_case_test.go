package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

func TestGetPaymentAddressStatusUseCaseSuccess(t *testing.T) {
	issuedAt := time.Date(2026, 3, 8, 11, 0, 0, 0, time.UTC)
	firstObservedAt := issuedAt.Add(10 * time.Minute)
	expiresAt := issuedAt.Add(24 * time.Hour)

	useCase := NewGetPaymentAddressStatusUseCase(
		&fakePaymentAddressStatusFinder{
			found: true,
			record: outport.PaymentAddressStatusRecord{
				PaymentAddressID:        101,
				AddressPolicyID:         "bitcoin-mainnet-native-segwit",
				ExpectedAmountMinor:     120000,
				CustomerReference:       "order-20260308-001",
				Chain:                   string(valueobjects.SupportedChainBitcoin),
				Network:                 string(valueobjects.NetworkIDMainnet),
				Scheme:                  string(valueobjects.AddressSchemeNativeSegwit),
				Address:                 "bc1qstatus",
				PaymentStatus:           string(valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted),
				ObservedTotalMinor:      80000,
				ConfirmedTotalMinor:     40000,
				UnconfirmedTotalMinor:   40000,
				RequiredConfirmations:   1,
				LastObservedBlockHeight: 123,
				IssuedAt:                issuedAt,
				FirstObservedAt:         &firstObservedAt,
				ExpiresAt:               &expiresAt,
				LastFailureReason:       string(valueobjects.PaymentReceiptTrackingFailureReasonObservationFailed),
			},
		},
		newInMemoryAddressPolicyReader([]policies.AddressIssuancePolicy{
			newAddressIssuancePolicy(
				"bitcoin-mainnet-native-segwit",
				valueobjects.SupportedChainBitcoin,
				valueobjects.NetworkIDMainnet,
				string(valueobjects.AddressSchemeNativeSegwit),
				8,
				"xpub",
				"m/84'/0'/0'",
			),
		}),
	)

	response, err := useCase.Execute(context.Background(), inport.GetPaymentAddressStatusInput{
		Chain:            string(valueobjects.SupportedChainBitcoin),
		PaymentAddressID: 101,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if response.PaymentAddressID != "101" {
		t.Fatalf("unexpected payment address id: got %q", response.PaymentAddressID)
	}
	if response.Decimals != 8 {
		t.Fatalf("unexpected decimals: got %d", response.Decimals)
	}
	if response.PaymentStatus != "paid_unconfirmed_reverted" {
		t.Fatalf("unexpected payment status: got %q", response.PaymentStatus)
	}
	if response.IssuedAt != issuedAt {
		t.Fatalf("unexpected issuedAt: got %v", response.IssuedAt)
	}
	if response.FirstObservedAt == nil || !response.FirstObservedAt.Equal(firstObservedAt) {
		t.Fatalf("unexpected firstObservedAt: got %v", response.FirstObservedAt)
	}
	if response.LastError != "receipt observation failed" {
		t.Fatalf("unexpected lastError: got %q", response.LastError)
	}
}

func TestGetPaymentAddressStatusUseCaseNotFound(t *testing.T) {
	useCase := NewGetPaymentAddressStatusUseCase(
		&fakePaymentAddressStatusFinder{},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), inport.GetPaymentAddressStatusInput{
		Chain:            string(valueobjects.SupportedChainBitcoin),
		PaymentAddressID: 404,
	})
	if !errors.Is(err, inport.ErrPaymentAddressNotFound) {
		t.Fatalf("expected ErrPaymentAddressNotFound, got %v", err)
	}
}

func TestGetPaymentAddressStatusUseCaseFinderError(t *testing.T) {
	useCase := NewGetPaymentAddressStatusUseCase(
		&fakePaymentAddressStatusFinder{err: outport.ErrPaymentAddressStatusIncomplete},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), inport.GetPaymentAddressStatusInput{
		Chain:            string(valueobjects.SupportedChainBitcoin),
		PaymentAddressID: 101,
	})
	if !errors.Is(err, inport.ErrInternalFailure) {
		t.Fatalf("expected ErrInternalFailure, got %v", err)
	}
}

func TestGetPaymentAddressStatusUseCaseMapsFinderDependencyFailure(t *testing.T) {
	useCase := NewGetPaymentAddressStatusUseCase(
		&fakePaymentAddressStatusFinder{err: errors.New("query failed")},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), inport.GetPaymentAddressStatusInput{
		Chain:            string(valueobjects.SupportedChainBitcoin),
		PaymentAddressID: 101,
	})
	if !errors.Is(err, inport.ErrDependencyFailure) {
		t.Fatalf("expected ErrDependencyFailure, got %v", err)
	}
}

func TestGetPaymentAddressStatusUseCasePolicyMissing(t *testing.T) {
	useCase := NewGetPaymentAddressStatusUseCase(
		&fakePaymentAddressStatusFinder{
			found: true,
			record: outport.PaymentAddressStatusRecord{
				PaymentAddressID: 101,
				AddressPolicyID:  "bitcoin-mainnet-native-segwit",
				Chain:            string(valueobjects.SupportedChainBitcoin),
				Network:          string(valueobjects.NetworkIDMainnet),
				Scheme:           string(valueobjects.AddressSchemeNativeSegwit),
				Address:          "bc1qstatus",
				PaymentStatus:    string(valueobjects.PaymentReceiptStatusWatching),
				IssuedAt:         time.Date(2026, 3, 8, 11, 0, 0, 0, time.UTC),
			},
		},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), inport.GetPaymentAddressStatusInput{
		Chain:            string(valueobjects.SupportedChainBitcoin),
		PaymentAddressID: 101,
	})
	if !errors.Is(err, inport.ErrPaymentAddressPolicyNotConfigured) {
		t.Fatalf("unexpected error: got %v", err)
	}
}
