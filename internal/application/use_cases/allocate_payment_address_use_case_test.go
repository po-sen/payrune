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
	"payrune/internal/domain/policies"
	"payrune/internal/domain/value_objects"
)

type fakeAllocatePaymentAddressClock struct {
	times []time.Time
	calls int
}

func newAllocatePaymentAddressClock() outport.Clock {
	return &fakeAllocatePaymentAddressClock{
		times: []time.Time{time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)},
	}
}

func (f *fakeAllocatePaymentAddressClock) NowUTC() time.Time {
	f.calls++
	if len(f.times) == 0 {
		return time.Time{}
	}
	if f.calls > len(f.times) {
		return f.times[len(f.times)-1]
	}
	return f.times[f.calls-1]
}

func newAllocateDeriveOutput(address string, path string) outport.DeriveChainAddressOutput {
	return outport.DeriveChainAddressOutput{
		Address:                address,
		RelativeDerivationPath: path,
	}
}

func newAllocationPolicy(
	addressPolicyID string,
	network value_objects.NetworkID,
	scheme string,
	publicKey string,
	publicKeyFingerprint string,
	derivationPathPrefix string,
) entities.AddressIssuancePolicy {
	return newAddressIssuancePolicy(
		addressPolicyID,
		value_objects.SupportedChainBitcoin,
		network,
		scheme,
		"satoshi",
		8,
		publicKey,
		testPublicKeyFingerprintAlgo,
		publicKeyFingerprint,
		derivationPathPrefix,
	)
}

func TestAllocatePaymentAddressUseCaseSuccess(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{}
	txManager := newFakeUnitOfWork(allocator)
	deriver := newFakeChainAddressDeriver()
	deriver.output = newAllocateDeriveOutput("bc1qallocatedaddress", "0/11")
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-native-segwit",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeNativeSegwit),
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
	useCase := NewAllocatePaymentAddressUseCase(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	response, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
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
	if allocator.lastReopenInput.IssuancePolicy.AddressPolicy.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf(
			"unexpected address policy id passed to allocator reopen: got %q",
			allocator.lastReopenInput.IssuancePolicy.AddressPolicy.AddressPolicyID,
		)
	}
	if allocator.lastReopenInput.IssuancePolicy.DerivationConfig.PublicKeyFingerprint == "" {
		t.Fatalf("expected non-empty xpub fingerprint")
	}
	if allocator.lastReopenInput.IssuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo != testPublicKeyFingerprintAlgo {
		t.Fatalf(
			"unexpected xpub fingerprint algorithm: got %q",
			allocator.lastReopenInput.IssuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo,
		)
	}
	if allocator.lastReopenInput.CustomerReference != "order-001" {
		t.Fatalf("unexpected customer reference passed to allocator reopen: got %q", allocator.lastReopenInput.CustomerReference)
	}
	if allocator.lastReopenInput.ExpectedAmountMinor != 120000 {
		t.Fatalf("unexpected expected amount minor passed to allocator reopen: got %d", allocator.lastReopenInput.ExpectedAmountMinor)
	}
	if allocator.lastReserveFreshInput.IssuancePolicy.AddressPolicy.AddressPolicyID != "bitcoin-mainnet-native-segwit" {
		t.Fatalf(
			"unexpected address policy id passed to allocator reserve fresh: got %q",
			allocator.lastReserveFreshInput.IssuancePolicy.AddressPolicy.AddressPolicyID,
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
	trackingStore, ok := txManager.receiptTrackingStore.(*fakeAllocatePaymentReceiptTrackingStore)
	if !ok {
		t.Fatal("expected fake receipt tracking store")
	}
	if trackingStore.createCalls != 1 {
		t.Fatalf("expected create tracking call count 1, got %d", trackingStore.createCalls)
	}
	if trackingStore.lastCreateTracking.PaymentAddressID != 44 {
		t.Fatalf(
			"unexpected payment address id passed to tracking create: got %d",
			trackingStore.lastCreateTracking.PaymentAddressID,
		)
	}
	if trackingStore.lastCreateTracking.RequiredConfirmations != 1 {
		t.Fatalf(
			"unexpected required confirmations passed to tracking create: got %d",
			trackingStore.lastCreateTracking.RequiredConfirmations,
		)
	}
	if trackingStore.lastCreateTracking.ExpiresAt == nil || trackingStore.lastCreateTracking.ExpiresAt.IsZero() {
		t.Fatal("expected non-zero expires at in created tracking")
	}
	if allocator.lastCompleteInput.PaymentAddressID != 44 {
		t.Fatalf("unexpected payment address id in complete input: got %d", allocator.lastCompleteInput.PaymentAddressID)
	}
	if allocator.lastCompleteInput.DerivationPath != "m/84'/0'/0'/0/11" {
		t.Fatalf("unexpected derivation path in complete input: got %q", allocator.lastCompleteInput.DerivationPath)
	}
	if deriver.lastInput.Index != 11 {
		t.Fatalf("unexpected index passed to deriver: got %d", deriver.lastInput.Index)
	}
	if deriver.lastInput.Network != value_objects.NetworkID(value_objects.BitcoinNetworkMainnet) {
		t.Fatalf("unexpected network passed to deriver: got %q", deriver.lastInput.Network)
	}
	if deriver.lastInput.Scheme != string(value_objects.BitcoinAddressSchemeNativeSegwit) {
		t.Fatalf("unexpected scheme passed to deriver: got %q", deriver.lastInput.Scheme)
	}
	if deriver.lastInput.AccountPublicKey != "xpub-main" {
		t.Fatalf("unexpected public key passed to deriver: got %q", deriver.lastInput.AccountPublicKey)
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

func TestAllocatePaymentAddressUseCaseUsesNetworkSpecificRequiredConfirmations(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{}
	txManager := newFakeUnitOfWork(allocator)
	deriver := newFakeChainAddressDeriver()
	deriver.output = newAllocateDeriveOutput("bc1qnetworkconfirmations", "0/15")
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-native-segwit",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeNativeSegwit),
			"xpub-main",
			"fingerprint-main-native-segwit",
			"m/84'/0'/0'",
		),
	})
	allocator.freshReservation = entities.PaymentAddressAllocation{
		PaymentAddressID:    66,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		DerivationIndex:     15,
		ExpectedAmountMinor: 25000,
		CustomerReference:   "order-66",
	}

	useCase := NewAllocatePaymentAddressUseCase(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(
			map[value_objects.NetworkID]int32{
				value_objects.NetworkID(value_objects.BitcoinNetworkMainnet):  6,
				value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4): 2,
			},
			nil,
		),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 25000,
		CustomerReference:   "order-66",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	trackingStore, ok := txManager.receiptTrackingStore.(*fakeAllocatePaymentReceiptTrackingStore)
	if !ok {
		t.Fatal("expected fake receipt tracking store")
	}
	if trackingStore.createCalls != 1 {
		t.Fatalf("expected create tracking call count 1, got %d", trackingStore.createCalls)
	}
	if trackingStore.lastCreateTracking.RequiredConfirmations != 6 {
		t.Fatalf("unexpected required confirmations: got %d", trackingStore.lastCreateTracking.RequiredConfirmations)
	}
}

func TestAllocatePaymentAddressUseCaseUsesNetworkSpecificReceiptExpiry(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{}
	txManager := newFakeUnitOfWork(allocator)
	deriver := newFakeChainAddressDeriver()
	deriver.output = newAllocateDeriveOutput("bc1qnetworkexpiry", "0/16")
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-native-segwit",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeNativeSegwit),
			"xpub-main",
			"fingerprint-main-native-segwit",
			"m/84'/0'/0'",
		),
	})
	allocator.freshReservation = entities.PaymentAddressAllocation{
		PaymentAddressID:    67,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		DerivationIndex:     16,
		ExpectedAmountMinor: 25000,
		CustomerReference:   "order-67",
	}
	now := time.Date(2026, 3, 6, 4, 0, 0, 0, time.UTC)
	clock := &fakeAllocatePaymentAddressClock{
		times: []time.Time{
			now,
			now.Add(6 * time.Hour),
		},
	}

	useCase := NewAllocatePaymentAddressUseCase(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(
			nil,
			map[value_objects.NetworkID]time.Duration{
				value_objects.NetworkID(value_objects.BitcoinNetworkMainnet):  48 * time.Hour,
				value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4): 24 * time.Hour,
			},
		),
		clock,
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 25000,
		CustomerReference:   "order-67",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	trackingStore, ok := txManager.receiptTrackingStore.(*fakeAllocatePaymentReceiptTrackingStore)
	if !ok {
		t.Fatal("expected fake receipt tracking store")
	}
	if trackingStore.createCalls != 1 {
		t.Fatalf("expected create tracking call count 1, got %d", trackingStore.createCalls)
	}
	expectedExpiresAt := now.Add(48 * time.Hour)
	if trackingStore.lastCreateTracking.ExpiresAt == nil || !trackingStore.lastCreateTracking.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected expires at: got %v, want %s", trackingStore.lastCreateTracking.ExpiresAt, expectedExpiresAt)
	}
	if clock.calls != 1 {
		t.Fatalf("expected clock to be read once, got %d", clock.calls)
	}
}

func TestAllocatePaymentAddressUseCaseReusesFailedReservationBeforeFresh(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{
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
	deriver := newFakeChainAddressDeriver()
	deriver.output = newAllocateDeriveOutput("bc1qreusedaddress", "0/7")
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-native-segwit",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeNativeSegwit),
			"xpub-main",
			"fingerprint-main-native-segwit",
			"m/84'/0'/0'",
		),
	})
	useCase := NewAllocatePaymentAddressUseCase(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	response, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
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
	trackingStore, ok := txManager.receiptTrackingStore.(*fakeAllocatePaymentReceiptTrackingStore)
	if !ok {
		t.Fatal("expected fake receipt tracking store")
	}
	if trackingStore.createCalls != 1 {
		t.Fatalf("expected create tracking call count 1, got %d", trackingStore.createCalls)
	}
	if trackingStore.lastCreateTracking.PaymentAddressID != 55 {
		t.Fatalf(
			"unexpected payment address id passed to tracking create: got %d",
			trackingStore.lastCreateTracking.PaymentAddressID,
		)
	}
	if trackingStore.lastCreateTracking.ExpiresAt == nil || trackingStore.lastCreateTracking.ExpiresAt.IsZero() {
		t.Fatal("expected non-zero expires at in created tracking")
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
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeLegacy),
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	allocator := &fakePaymentAddressAllocationStore{
		freshReservation: entities.PaymentAddressAllocation{
			PaymentAddressID:    77,
			AddressPolicyID:     "bitcoin-mainnet-legacy",
			DerivationIndex:     2,
			ExpectedAmountMinor: 1,
		},
	}
	txManager := &fakeUnitOfWork{
		err:             expectedErr,
		allocationStore: allocator,
	}
	useCase := NewAllocatePaymentAddressUseCase(
		txManager,
		newFakeChainAddressDeriver(),
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
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

func TestAllocatePaymentAddressUseCaseReturnsTrackingRegistrationError(t *testing.T) {
	expectedErr := errors.New("register tracking failed")
	allocator := &fakePaymentAddressAllocationStore{
		freshReservation: entities.PaymentAddressAllocation{
			PaymentAddressID:    88,
			AddressPolicyID:     "bitcoin-mainnet-native-segwit",
			DerivationIndex:     4,
			ExpectedAmountMinor: 500,
		},
	}
	txManager := newFakeUnitOfWork(allocator)
	trackingStore, ok := txManager.receiptTrackingStore.(*fakeAllocatePaymentReceiptTrackingStore)
	if !ok {
		t.Fatal("expected fake receipt tracking store")
	}
	trackingStore.createErr = expectedErr
	deriver := newFakeChainAddressDeriver()
	deriver.output = newAllocateDeriveOutput("bc1qtrackingerror", "0/4")
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-native-segwit",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeNativeSegwit),
			"xpub-main",
			"fingerprint-main-native-segwit",
			"m/84'/0'/0'",
		),
	})
	useCase := NewAllocatePaymentAddressUseCase(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 500,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if allocator.completeCalls != 1 {
		t.Fatalf("expected complete allocation call count 1, got %d", allocator.completeCalls)
	}
	if trackingStore.createCalls != 1 {
		t.Fatalf("expected create tracking call count 1, got %d", trackingStore.createCalls)
	}
}

func TestAllocatePaymentAddressUseCaseRejectUnsupportedChain(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		newFakeChainAddressDeriver(),
		newInMemoryAddressPolicyReader(nil),
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChain("eth"),
		AddressPolicyID:     "eth-mainnet",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, inport.ErrChainNotSupported) {
		t.Fatalf("expected chain not supported error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseRejectUnknownPolicy(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		newFakeChainAddressDeriver(),
		newInMemoryAddressPolicyReader(nil),
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, inport.ErrAddressPolicyNotFound) {
		t.Fatalf("expected address policy not found error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseRejectDisabledPolicy(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAddressIssuancePolicy(
			"bitcoin-mainnet-legacy",
			value_objects.SupportedChainBitcoin,
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeLegacy),
			"satoshi",
			8,
			"",
			"",
			"",
			"",
		),
	})
	allocator := &fakePaymentAddressAllocationStore{}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		newFakeChainAddressDeriver(),
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, inport.ErrAddressPolicyNotEnabled) {
		t.Fatalf("expected address policy not enabled error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseMapsExhaustedError(t *testing.T) {
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeLegacy),
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	allocator := &fakePaymentAddressAllocationStore{reserveFreshErr: outport.ErrAddressIndexExhausted}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		newFakeChainAddressDeriver(),
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, inport.ErrAddressPoolExhausted) {
		t.Fatalf("expected address pool exhausted error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseDerivationError(t *testing.T) {
	expectedErr := errors.New("derive failed")
	allocator := &fakePaymentAddressAllocationStore{
		freshReservation: entities.PaymentAddressAllocation{
			PaymentAddressID:    99,
			AddressPolicyID:     "bitcoin-mainnet-legacy",
			DerivationIndex:     1,
			ExpectedAmountMinor: 1,
		},
	}
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeLegacy),
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		&fakeChainAddressDeriver{
			supportedChains: map[value_objects.SupportedChain]bool{
				value_objects.SupportedChainBitcoin: true,
			},
			err: expectedErr,
		},
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
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
	allocator := &fakePaymentAddressAllocationStore{
		freshReservation: entities.PaymentAddressAllocation{
			PaymentAddressID:    88,
			AddressPolicyID:     "bitcoin-mainnet-legacy",
			DerivationIndex:     3,
			ExpectedAmountMinor: 1,
		},
	}
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeLegacy),
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		&fakeChainAddressDeriver{
			supportedChains: map[value_objects.SupportedChain]bool{
				value_objects.SupportedChainBitcoin: true,
			},
			err: expectedErr,
		},
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
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
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-legacy",
			value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			string(value_objects.BitcoinAddressSchemeLegacy),
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	allocator := &fakePaymentAddressAllocationStore{}
	useCase := NewAllocatePaymentAddressUseCase(
		newFakeUnitOfWork(allocator),
		newFakeChainAddressDeriver(),
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 0,
	})
	if !errors.Is(err, inport.ErrInvalidExpectedAmount) {
		t.Fatalf("expected invalid expected amount error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseValidationMissingDependencies(t *testing.T) {
	input := dto.AllocatePaymentAddressInput{
		Chain:               value_objects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
	}

	tests := []struct {
		name    string
		useCase *allocatePaymentAddressUseCase
		wantErr string
	}{
		{
			name:    "missing unit of work",
			useCase: &allocatePaymentAddressUseCase{},
			wantErr: "unit of work is not configured",
		},
		{
			name: "missing deriver",
			useCase: &allocatePaymentAddressUseCase{
				unitOfWork:   newFakeUnitOfWork(&fakePaymentAddressAllocationStore{}),
				policyReader: newInMemoryAddressPolicyReader(nil),
				clock:        newAllocatePaymentAddressClock(),
			},
			wantErr: "chain address deriver is not configured",
		},
		{
			name: "missing policy reader",
			useCase: &allocatePaymentAddressUseCase{
				unitOfWork: newFakeUnitOfWork(&fakePaymentAddressAllocationStore{}),
				deriver:    newFakeChainAddressDeriver(),
				clock:      newAllocatePaymentAddressClock(),
			},
			wantErr: "address policy reader is not configured",
		},
		{
			name: "missing clock",
			useCase: &allocatePaymentAddressUseCase{
				unitOfWork:   newFakeUnitOfWork(&fakePaymentAddressAllocationStore{}),
				deriver:      newFakeChainAddressDeriver(),
				policyReader: newInMemoryAddressPolicyReader(nil),
			},
			wantErr: "clock is not configured",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.useCase.Execute(context.Background(), input)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("unexpected error: got %v want %q", err, tc.wantErr)
			}
		})
	}
}
