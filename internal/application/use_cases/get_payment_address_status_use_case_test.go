package use_cases

import (
	"context"
	"errors"
	"testing"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
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
				Chain:                   value_objects.SupportedChainBitcoin,
				Network:                 value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
				Scheme:                  string(value_objects.BitcoinAddressSchemeNativeSegwit),
				Address:                 "bc1qstatus",
				PaymentStatus:           value_objects.PaymentReceiptStatusPaidUnconfirmedReverted,
				ObservedTotalMinor:      80000,
				ConfirmedTotalMinor:     40000,
				UnconfirmedTotalMinor:   40000,
				RequiredConfirmations:   1,
				LastObservedBlockHeight: 123,
				IssuedAt:                issuedAt,
				FirstObservedAt:         &firstObservedAt,
				ExpiresAt:               &expiresAt,
			},
		},
		newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
			newAddressIssuancePolicy(
				"bitcoin-mainnet-native-segwit",
				value_objects.SupportedChainBitcoin,
				value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
				string(value_objects.BitcoinAddressSchemeNativeSegwit),
				"satoshi",
				8,
				"xpub",
				testPublicKeyFingerprintAlgo,
				"fingerprint",
				"m/84'/0'/0'",
			),
		}),
	)

	response, err := useCase.Execute(context.Background(), dto.GetPaymentAddressStatusInput{
		Chain:            value_objects.SupportedChainBitcoin,
		PaymentAddressID: 101,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if response.PaymentAddressID != "101" {
		t.Fatalf("unexpected payment address id: got %q", response.PaymentAddressID)
	}
	if response.MinorUnit != "satoshi" {
		t.Fatalf("unexpected minor unit: got %q", response.MinorUnit)
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
}

func TestGetPaymentAddressStatusUseCaseNotFound(t *testing.T) {
	useCase := NewGetPaymentAddressStatusUseCase(
		&fakePaymentAddressStatusFinder{},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), dto.GetPaymentAddressStatusInput{
		Chain:            value_objects.SupportedChainBitcoin,
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

	_, err := useCase.Execute(context.Background(), dto.GetPaymentAddressStatusInput{
		Chain:            value_objects.SupportedChainBitcoin,
		PaymentAddressID: 101,
	})
	if !errors.Is(err, outport.ErrPaymentAddressStatusIncomplete) {
		t.Fatalf("expected ErrPaymentAddressStatusIncomplete, got %v", err)
	}
}

func TestGetPaymentAddressStatusUseCasePolicyMissing(t *testing.T) {
	useCase := NewGetPaymentAddressStatusUseCase(
		&fakePaymentAddressStatusFinder{
			found: true,
			record: outport.PaymentAddressStatusRecord{
				PaymentAddressID: 101,
				AddressPolicyID:  "bitcoin-mainnet-native-segwit",
				Chain:            value_objects.SupportedChainBitcoin,
				Network:          value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
				Scheme:           string(value_objects.BitcoinAddressSchemeNativeSegwit),
				Address:          "bc1qstatus",
				PaymentStatus:    value_objects.PaymentReceiptStatusWatching,
				IssuedAt:         time.Date(2026, 3, 8, 11, 0, 0, 0, time.UTC),
			},
		},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), dto.GetPaymentAddressStatusInput{
		Chain:            value_objects.SupportedChainBitcoin,
		PaymentAddressID: 101,
	})
	if err == nil || err.Error() != "payment address policy is not configured" {
		t.Fatalf("unexpected error: got %v", err)
	}
}
