package use_cases

import (
	"context"
	"errors"
	"testing"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

func newAllocationPolicy(
	addressPolicyID string,
	network value_objects.BitcoinNetwork,
	scheme value_objects.BitcoinAddressScheme,
	xpub string,
	xpubFingerprint string,
	derivationPathPrefix string,
) entities.AddressPolicy {
	return entities.AddressPolicy{
		AddressPolicyID:      addressPolicyID,
		Chain:                value_objects.ChainBitcoin,
		Network:              network,
		Scheme:               scheme,
		MinorUnit:            "satoshi",
		Decimals:             8,
		XPub:                 xpub,
		XPubFingerprintAlgo:  testXPubFingerprintAlgo,
		XPubFingerprint:      xpubFingerprint,
		DerivationPathPrefix: derivationPathPrefix,
	}
}

func TestAllocatePaymentAddressUseCaseSuccess(t *testing.T) {
	allocator := &fakePaymentAddressAllocationRepository{}
	txManager := newFakeUnitOfWork(allocator)
	deriver := &fakePolicyBitcoinAddressDeriver{
		address:        "bc1qallocatedaddress",
		derivationPath: "0/11",
	}
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-native-segwit",
			value_objects.BitcoinNetworkMainnet,
			value_objects.BitcoinAddressSchemeNativeSegwit,
			"xpub-main",
			"fingerprint-main-native-segwit",
			"m/84'/0'/0'",
		),
	})
	allocator.freshReservation = entities.PaymentAddressAllocation{
		PaymentAddressID:    44,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		DerivationIndex:     11,
		ExpectedAmountMinor: 120000,
		CustomerReference:   "order-001",
	}
	useCase := NewAllocatePaymentAddressUseCase(txManager, deriver, catalog)

	response, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.ChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 120000,
		CustomerReference:   " order-001 ",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if allocator.reopenCalls != 1 {
		t.Fatalf("expected reopen failed reservation call count 1, got %d", allocator.reopenCalls)
	}
	if txManager.calls != 1 {
		t.Fatalf("expected transaction manager call count 1, got %d", txManager.calls)
	}
	if allocator.reserveFreshCalls != 1 {
		t.Fatalf("expected reserve fresh index call count 1, got %d", allocator.reserveFreshCalls)
	}
	if allocator.lastReopenInput.Policy.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf(
			"unexpected address policy id passed to allocator reopen: got %q",
			allocator.lastReopenInput.Policy.AddressPolicyID,
		)
	}
	if allocator.lastReopenInput.Policy.XPubFingerprint == "" {
		t.Fatalf("expected non-empty xpub fingerprint")
	}
	if allocator.lastReopenInput.Policy.XPubFingerprintAlgo != testXPubFingerprintAlgo {
		t.Fatalf(
			"unexpected xpub fingerprint algorithm: got %q",
			allocator.lastReopenInput.Policy.XPubFingerprintAlgo,
		)
	}
	if allocator.lastReopenInput.CustomerReference != "order-001" {
		t.Fatalf("unexpected customer reference passed to allocator reopen: got %q", allocator.lastReopenInput.CustomerReference)
	}
	if allocator.lastReopenInput.ExpectedAmountMinor != 120000 {
		t.Fatalf("unexpected expected amount minor passed to allocator reopen: got %d", allocator.lastReopenInput.ExpectedAmountMinor)
	}
	if allocator.lastReserveFreshInput.Policy.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf(
			"unexpected address policy id passed to allocator reserve fresh: got %q",
			allocator.lastReserveFreshInput.Policy.AddressPolicyID,
		)
	}
	if allocator.lastReserveFreshInput.CustomerReference != "order-001" {
		t.Fatalf("unexpected customer reference passed to allocator reserve fresh: got %q", allocator.lastReserveFreshInput.CustomerReference)
	}
	if allocator.lastReserveFreshInput.ExpectedAmountMinor != 120000 {
		t.Fatalf("unexpected expected amount minor passed to allocator reserve fresh: got %d", allocator.lastReserveFreshInput.ExpectedAmountMinor)
	}
	if allocator.completeCalls != 1 {
		t.Fatalf("expected complete allocation call count 1, got %d", allocator.completeCalls)
	}
	if allocator.lastCompleteInput.PaymentAddressID != 44 {
		t.Fatalf("unexpected payment address id in complete input: got %d", allocator.lastCompleteInput.PaymentAddressID)
	}
	if allocator.lastCompleteInput.DerivationPath != "m/84'/0'/0'/0/11" {
		t.Fatalf("unexpected derivation path in complete input: got %q", allocator.lastCompleteInput.DerivationPath)
	}
	if deriver.lastIndex != 11 {
		t.Fatalf("unexpected index passed to deriver: got %d", deriver.lastIndex)
	}
	if deriver.lastDerivationIndex != 11 {
		t.Fatalf("unexpected index passed to derivation path builder: got %d", deriver.lastDerivationIndex)
	}
	if response.Address != "bc1qallocatedaddress" {
		t.Fatalf("unexpected address: got %q", response.Address)
	}
	if response.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf("unexpected address policy id: got %q", response.AddressPolicyID)
	}
	if response.PaymentAddressID != "44" {
		t.Fatalf("unexpected payment address id: got %q", response.PaymentAddressID)
	}
	if response.ExpectedAmountMinor != 120000 {
		t.Fatalf("unexpected expected amount minor: got %d", response.ExpectedAmountMinor)
	}
	if response.CustomerReference != "order-001" {
		t.Fatalf("unexpected customer reference: got %q", response.CustomerReference)
	}
}

func TestAllocatePaymentAddressUseCaseReusesFailedReservationBeforeFresh(t *testing.T) {
	allocator := &fakePaymentAddressAllocationRepository{
		reopenFound: true,
		reopenReservation: entities.PaymentAddressAllocation{
			PaymentAddressID:    55,
			AddressPolicyID:     "bitcoin-mainnet-native-segwit",
			DerivationIndex:     7,
			ExpectedAmountMinor: 5000,
			CustomerReference:   "invoice-55",
		},
	}
	txManager := newFakeUnitOfWork(allocator)
	deriver := &fakePolicyBitcoinAddressDeriver{
		address:        "bc1qreusedaddress",
		derivationPath: "0/7",
	}
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-native-segwit",
			value_objects.BitcoinNetworkMainnet,
			value_objects.BitcoinAddressSchemeNativeSegwit,
			"xpub-main",
			"fingerprint-main-native-segwit",
			"m/84'/0'/0'",
		),
	})
	useCase := NewAllocatePaymentAddressUseCase(txManager, deriver, catalog)

	response, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.ChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 5000,
		CustomerReference:   "invoice-55",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if allocator.reopenCalls != 1 {
		t.Fatalf("expected reopen failed reservation call count 1, got %d", allocator.reopenCalls)
	}
	if txManager.calls != 1 {
		t.Fatalf("expected transaction manager call count 1, got %d", txManager.calls)
	}
	if allocator.reserveFreshCalls != 0 {
		t.Fatalf("expected reserve fresh index call count 0, got %d", allocator.reserveFreshCalls)
	}
	if allocator.completeCalls != 1 {
		t.Fatalf("expected complete allocation call count 1, got %d", allocator.completeCalls)
	}
	if allocator.lastCompleteInput.PaymentAddressID != 55 {
		t.Fatalf("unexpected payment address id in complete input: got %d", allocator.lastCompleteInput.PaymentAddressID)
	}
	if response.PaymentAddressID != "55" {
		t.Fatalf("unexpected payment address id: got %q", response.PaymentAddressID)
	}
}

func TestAllocatePaymentAddressUseCaseReturnsTransactionError(t *testing.T) {
	expectedErr := errors.New("transaction failed")
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.BitcoinNetworkMainnet,
			value_objects.BitcoinAddressSchemeLegacy,
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	allocator := &fakePaymentAddressAllocationRepository{
		freshReservation: entities.PaymentAddressAllocation{
			PaymentAddressID:    77,
			AddressPolicyID:     "bitcoin-mainnet-legacy",
			DerivationIndex:     2,
			ExpectedAmountMinor: 1,
		},
	}
	txManager := &fakeUnitOfWork{
		err:        expectedErr,
		repository: allocator,
	}
	useCase := NewAllocatePaymentAddressUseCase(
		txManager,
		&fakePolicyBitcoinAddressDeriver{},
		catalog,
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.ChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if allocator.reopenCalls != 0 {
		t.Fatalf("expected reopen not to be called when transaction manager fails, got %d", allocator.reopenCalls)
	}
}

func TestAllocatePaymentAddressUseCaseRejectUnsupportedChain(t *testing.T) {
	allocator := &fakePaymentAddressAllocationRepository{}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		&fakePolicyBitcoinAddressDeriver{},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.Chain("eth"),
		AddressPolicyID:     "eth-mainnet",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, inport.ErrChainNotSupported) {
		t.Fatalf("expected chain not supported error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseRejectUnknownPolicy(t *testing.T) {
	allocator := &fakePaymentAddressAllocationRepository{}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		&fakePolicyBitcoinAddressDeriver{},
		newInMemoryAddressPolicyReader(nil),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.ChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, inport.ErrAddressPolicyNotFound) {
		t.Fatalf("expected address policy not found error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseRejectDisabledPolicy(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		{
			AddressPolicyID: "bitcoin-mainnet-legacy",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeLegacy,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            "",
		},
	})
	allocator := &fakePaymentAddressAllocationRepository{}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		&fakePolicyBitcoinAddressDeriver{},
		catalog,
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.ChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, inport.ErrAddressPolicyNotEnabled) {
		t.Fatalf("expected address policy not enabled error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseMapsExhaustedError(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.BitcoinNetworkMainnet,
			value_objects.BitcoinAddressSchemeLegacy,
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	allocator := &fakePaymentAddressAllocationRepository{reserveFreshErr: outport.ErrAddressIndexExhausted}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		&fakePolicyBitcoinAddressDeriver{},
		catalog,
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.ChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, inport.ErrAddressPoolExhausted) {
		t.Fatalf("expected address pool exhausted error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseDerivationError(t *testing.T) {
	expectedErr := errors.New("derive failed")
	allocator := &fakePaymentAddressAllocationRepository{
		freshReservation: entities.PaymentAddressAllocation{
			PaymentAddressID:    99,
			AddressPolicyID:     "bitcoin-mainnet-legacy",
			DerivationIndex:     1,
			ExpectedAmountMinor: 1,
		},
	}
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.BitcoinNetworkMainnet,
			value_objects.BitcoinAddressSchemeLegacy,
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		&fakePolicyBitcoinAddressDeriver{err: expectedErr},
		catalog,
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.ChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if allocator.markFailedCalls != 1 {
		t.Fatalf("expected mark failed to be called once, got %d", allocator.markFailedCalls)
	}
	if allocator.lastFailedInput.PaymentAddressID != 99 {
		t.Fatalf("unexpected failed payment address id: got %d", allocator.lastFailedInput.PaymentAddressID)
	}
	if allocator.lastFailedInput.FailureReason == "" {
		t.Fatalf("expected non-empty failure reason")
	}
	if allocator.lastFailedInput.Status != value_objects.PaymentAddressAllocationStatusDerivationFailed {
		t.Fatalf("unexpected failed status: got %q", allocator.lastFailedInput.Status)
	}
	if allocator.completeCalls != 0 {
		t.Fatalf("expected complete allocation not to be called on derivation error")
	}
}

func TestAllocatePaymentAddressUseCaseDerivationPathError(t *testing.T) {
	expectedErr := errors.New("path failed")
	allocator := &fakePaymentAddressAllocationRepository{
		freshReservation: entities.PaymentAddressAllocation{
			PaymentAddressID:    88,
			AddressPolicyID:     "bitcoin-mainnet-legacy",
			DerivationIndex:     3,
			ExpectedAmountMinor: 1,
		},
	}
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.BitcoinNetworkMainnet,
			value_objects.BitcoinAddressSchemeLegacy,
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		&fakePolicyBitcoinAddressDeriver{address: "1BitcoinAddressExample", derivationPathErr: expectedErr},
		catalog,
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.ChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if allocator.markFailedCalls != 1 {
		t.Fatalf("expected mark failed to be called once, got %d", allocator.markFailedCalls)
	}
	if allocator.completeCalls != 0 {
		t.Fatalf("expected complete allocation not to be called when derivation path fails")
	}
}

func TestAllocatePaymentAddressUseCaseRejectInvalidExpectedAmount(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressPolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.BitcoinNetworkMainnet,
			value_objects.BitcoinAddressSchemeLegacy,
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	allocator := &fakePaymentAddressAllocationRepository{}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		&fakePolicyBitcoinAddressDeriver{},
		catalog,
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.ChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 0,
	})
	if !errors.Is(err, inport.ErrInvalidExpectedAmount) {
		t.Fatalf("expected invalid expected amount error, got %v", err)
	}
}
