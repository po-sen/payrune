package usecases

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
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
	network valueobjects.NetworkID,
	scheme string,
	publicKey string,
	publicKeyFingerprint string,
	derivationPathPrefix string,
) entities.AddressIssuancePolicy {
	return newAddressIssuancePolicy(
		addressPolicyID,
		valueobjects.SupportedChainBitcoin,
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

func newAllocatePaymentAddressUseCaseForTest(
	txManager *fakeUnitOfWork,
	deriver outport.ChainAddressDeriver,
	policyReader outport.AddressPolicyReader,
	issuancePolicy policies.PaymentAddressAllocationIssuancePolicy,
	clock outport.Clock,
) inport.AllocatePaymentAddressUseCase {
	return NewAllocatePaymentAddressUseCase(
		txManager,
		deriver,
		policyReader,
		issuancePolicy,
		clock,
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
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
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	response, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 120000,
		CustomerReference:   " order-001 ",
		IdempotencyKey:      " idem-001 ",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if allocator.reopenCalls != 1 {
		t.Fatalf("expected reopen failed reservation call count 1, got %d", allocator.reopenCalls)
	}
	if txManager.calls != 2 {
		t.Fatalf("expected transaction manager call count 2, got %d", txManager.calls)
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
	idempotencyStore, ok := txManager.idempotencyStore.(*fakePaymentAddressIdempotencyStore)
	if !ok {
		t.Fatal("expected fake idempotency store")
	}
	if idempotencyStore.claimCalls != 1 {
		t.Fatalf("expected idempotency claim call count 1, got %d", idempotencyStore.claimCalls)
	}
	if idempotencyStore.lastClaim.IdempotencyKey != "idem-001" {
		t.Fatalf("unexpected idempotency key in claim input: got %q", idempotencyStore.lastClaim.IdempotencyKey)
	}
	if idempotencyStore.completeCalls != 1 {
		t.Fatalf("expected idempotency complete call count 1, got %d", idempotencyStore.completeCalls)
	}
	if idempotencyStore.lastComplete.PaymentAddressID != 44 {
		t.Fatalf("unexpected payment address id in idempotency complete input: got %d", idempotencyStore.lastComplete.PaymentAddressID)
	}
	if idempotencyStore.releaseCalls != 0 {
		t.Fatalf("expected idempotency release not to be called, got %d", idempotencyStore.releaseCalls)
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
	if deriver.lastInput.Network != valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet) {
		t.Fatalf("unexpected network passed to deriver: got %q", deriver.lastInput.Network)
	}
	if deriver.lastInput.Scheme != string(valueobjects.BitcoinAddressSchemeNativeSegwit) {
		t.Fatalf("unexpected scheme passed to deriver: got %q", deriver.lastInput.Scheme)
	}
	if deriver.lastInput.AccountPublicKey != "xpub-main" {
		t.Fatalf("unexpected public key passed to deriver: got %q", deriver.lastInput.AccountPublicKey)
	}
	if deriver.lastInput.DerivationPathPrefix != "m/84'/0'/0'" {
		t.Fatalf("unexpected derivation path prefix passed to deriver: got %q", deriver.lastInput.DerivationPathPrefix)
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
	if response.IdempotencyReplayed {
		t.Fatal("expected fresh allocation response not to be marked as replayed")
	}
}

func TestAllocatePaymentAddressUseCaseEthereumSuccess(t *testing.T) {
	previousSaltReader := ethereumSaltReader
	ethereumSaltReader = bytes.NewReader(bytes.Repeat([]byte{0x11}, 32))
	defer func() {
		ethereumSaltReader = previousSaltReader
	}()

	allocator := &fakePaymentAddressAllocationStore{}
	txManager := newFakeUnitOfWork(allocator)
	txManager.evmFactoryStore = &fakeEVMFactoryStore{
		found: true,
		record: outport.EVMFactoryRecord{
			ID:                    7,
			Network:               valueobjects.NetworkID("sepolia"),
			FactoryAddress:        "0x3333333333333333333333333333333333333333",
			CollectorAddress:      "0x4444444444444444444444444444444444444444",
			VaultCreationCodeHash: "0x4023b4c6529dc6ac7f0b3db124a2b8c86febf4ab1e6e491612ca4a068fae6e21",
		},
	}
	txManager.evmVaultStore = &fakeEVMPaymentVaultStore{}

	deriver := newFakeChainAddressDeriver()
	deriver.supportedChains[valueobjects.SupportedChainEthereum] = true

	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		(entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "ethereum-sepolia-eth",
				Chain:           valueobjects.SupportedChainEthereum,
				Network:         valueobjects.NetworkID("sepolia"),
				Scheme:          "create2_forwarder",
				AssetCode:       "eth",
				AssetType:       "native",
				MinorUnit:       "wei",
				Decimals:        18,
			},
			DerivationConfig: valueobjects.AddressDerivationConfig{
				AccountPublicKey:         "0xplaceholder",
				PublicKeyFingerprintAlgo: testPublicKeyFingerprintAlgo,
				PublicKeyFingerprint:     "placeholder-fingerprint",
			},
		}).Normalize(),
	})
	allocator.freshReservation = entities.PaymentAddressAllocation{
		PaymentAddressID:    77,
		AddressPolicyID:     "ethereum-sepolia-eth",
		DerivationIndex:     0,
		ExpectedAmountMinor: 123,
		CustomerReference:   "eth-order-001",
	}
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(
			map[valueobjects.NetworkID]int32{
				valueobjects.NetworkID("sepolia"): 12,
			},
			map[valueobjects.NetworkID]time.Duration{
				valueobjects.NetworkID("sepolia"): time.Hour,
			},
		),
		newAllocatePaymentAddressClock(),
	)

	response, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainEthereum,
		AddressPolicyID:     "ethereum-sepolia-eth",
		ExpectedAmountMinor: 123,
		CustomerReference:   " eth-order-001 ",
		IdempotencyKey:      " eth-idem-001 ",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	expectedAddress := "0x5816b1fbecac478596e7436d2cee2cf37071f47b"
	if response.Address != expectedAddress {
		t.Fatalf("unexpected ethereum address: got %q", response.Address)
	}
	if response.AssetCode != "eth" {
		t.Fatalf("unexpected asset code: got %q", response.AssetCode)
	}
	if response.AssetType != "native" {
		t.Fatalf("unexpected asset type: got %q", response.AssetType)
	}
	if response.MinorUnit != "wei" {
		t.Fatalf("unexpected minor unit: got %q", response.MinorUnit)
	}
	if response.Decimals != 18 {
		t.Fatalf("unexpected decimals: got %d", response.Decimals)
	}
	if allocator.reserveFreshCalls != 1 {
		t.Fatalf("expected reserve fresh call count 1, got %d", allocator.reserveFreshCalls)
	}
	if allocator.completeCalls != 1 {
		t.Fatalf("expected complete call count 1, got %d", allocator.completeCalls)
	}
	if allocator.lastReserveFreshInput.IssuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo != evmFactoryAddressFingerprintAlgorithm {
		t.Fatalf(
			"unexpected ethereum fingerprint algorithm: got %q",
			allocator.lastReserveFreshInput.IssuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo,
		)
	}
	if allocator.lastReserveFreshInput.IssuancePolicy.DerivationConfig.PublicKeyFingerprint != strings.ToLower("0x3333333333333333333333333333333333333333") {
		t.Fatalf(
			"unexpected ethereum fingerprint: got %q",
			allocator.lastReserveFreshInput.IssuancePolicy.DerivationConfig.PublicKeyFingerprint,
		)
	}
	if allocator.lastCompleteInput.Address != expectedAddress {
		t.Fatalf("unexpected complete address: got %q", allocator.lastCompleteInput.Address)
	}
	if allocator.lastCompleteInput.IssuanceMethod != "create2_forwarder" {
		t.Fatalf("unexpected issuance method: got %q", allocator.lastCompleteInput.IssuanceMethod)
	}
	if allocator.lastCompleteInput.DerivationPath != "" {
		t.Fatalf("unexpected derivation path: got %q", allocator.lastCompleteInput.DerivationPath)
	}
	if deriver.calls != 0 {
		t.Fatalf("expected bitcoin deriver not to be called, got %d", deriver.calls)
	}

	vaultStore, ok := txManager.evmVaultStore.(*fakeEVMPaymentVaultStore)
	if !ok {
		t.Fatal("expected fake evm payment vault store")
	}
	if vaultStore.createCalls != 1 {
		t.Fatalf("expected evm vault create call count 1, got %d", vaultStore.createCalls)
	}
	if vaultStore.lastCreateInput.PaymentAddressID != 77 {
		t.Fatalf("unexpected vault payment address id: got %d", vaultStore.lastCreateInput.PaymentAddressID)
	}
	if vaultStore.lastCreateInput.FactoryID != 7 {
		t.Fatalf("unexpected vault factory id: got %d", vaultStore.lastCreateInput.FactoryID)
	}
	if vaultStore.lastCreateInput.PredictedAddress != expectedAddress {
		t.Fatalf("unexpected predicted address: got %q", vaultStore.lastCreateInput.PredictedAddress)
	}
	if vaultStore.lastCreateInput.SaltHex != "0x1111111111111111111111111111111111111111111111111111111111111111" {
		t.Fatalf("unexpected salt hex: got %q", vaultStore.lastCreateInput.SaltHex)
	}

	trackingStore, ok := txManager.receiptTrackingStore.(*fakeAllocatePaymentReceiptTrackingStore)
	if !ok {
		t.Fatal("expected fake receipt tracking store")
	}
	if trackingStore.createCalls != 1 {
		t.Fatalf("expected receipt tracking create call count 1, got %d", trackingStore.createCalls)
	}
	if trackingStore.lastCreateTracking.Address != expectedAddress {
		t.Fatalf("unexpected tracking address: got %q", trackingStore.lastCreateTracking.Address)
	}
	if trackingStore.lastCreateTracking.AssetCode != "eth" {
		t.Fatalf("unexpected tracking asset code: got %q", trackingStore.lastCreateTracking.AssetCode)
	}
	if trackingStore.lastCreateTracking.RequiredConfirmations != 12 {
		t.Fatalf("unexpected tracking confirmations: got %d", trackingStore.lastCreateTracking.RequiredConfirmations)
	}
	if trackingStore.lastCreateTracking.ExpiresAt == nil || !trackingStore.lastCreateTracking.ExpiresAt.Equal(time.Date(2026, 3, 6, 1, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected tracking expiresAt: got %+v", trackingStore.lastCreateTracking.ExpiresAt)
	}
}

func TestAllocatePaymentAddressUseCaseRejectsEthereumPolicyWithoutActiveFactory(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{}
	txManager := newFakeUnitOfWork(allocator)
	txManager.evmFactoryStore = &fakeEVMFactoryStore{found: false}
	txManager.evmVaultStore = &fakeEVMPaymentVaultStore{}

	deriver := newFakeChainAddressDeriver()
	deriver.supportedChains[valueobjects.SupportedChainEthereum] = true

	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		(entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "ethereum-sepolia-eth",
				Chain:           valueobjects.SupportedChainEthereum,
				Network:         valueobjects.NetworkID("sepolia"),
				Scheme:          "create2_forwarder",
				AssetCode:       "eth",
				AssetType:       "native",
				MinorUnit:       "wei",
				Decimals:        18,
			},
			DerivationConfig: valueobjects.AddressDerivationConfig{
				AccountPublicKey:         "0xplaceholder",
				PublicKeyFingerprintAlgo: testPublicKeyFingerprintAlgo,
				PublicKeyFingerprint:     "placeholder-fingerprint",
			},
		}).Normalize(),
	})

	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainEthereum,
		AddressPolicyID:     "ethereum-sepolia-eth",
		ExpectedAmountMinor: 123,
	})
	if !errors.Is(err, inport.ErrAddressPolicyNotEnabled) {
		t.Fatalf("expected ErrAddressPolicyNotEnabled, got %v", err)
	}
	if allocator.reserveFreshCalls != 0 {
		t.Fatalf("expected no reserve fresh calls, got %d", allocator.reserveFreshCalls)
	}
	if deriver.calls != 0 {
		t.Fatalf("expected deriver not to be called, got %d", deriver.calls)
	}
}

func TestAllocatePaymentAddressUseCaseReturnsExistingIssuedAllocationForDuplicateIdempotencyKey(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{
		issuedByIDFound: true,
		issuedByID: entities.PaymentAddressAllocation{
			PaymentAddressID:    71,
			AddressPolicyID:     "bitcoin-mainnet-native-segwit",
			DerivationIndex:     9,
			ExpectedAmountMinor: 120000,
			CustomerReference:   "order-duplicate",
			Status:              valueobjects.PaymentAddressAllocationStatusIssued,
			Chain:               valueobjects.SupportedChainBitcoin,
			Network:             valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:              string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			Address:             "bc1qexistingduplicate",
			DerivationPath:      "m/84'/0'/0'/0/9",
		},
	}
	idempotencyStore := &fakePaymentAddressIdempotencyStore{
		found: true,
		record: outport.PaymentAddressIdempotencyRecord{
			Chain:               valueobjects.SupportedChainBitcoin,
			IdempotencyKey:      "idem-duplicate",
			AddressPolicyID:     "bitcoin-mainnet-native-segwit",
			ExpectedAmountMinor: 120000,
			CustomerReference:   "order-duplicate",
			PaymentAddressID:    71,
		},
	}
	txManager := newFakeUnitOfWork(allocator)
	txManager.idempotencyStore = idempotencyStore
	deriver := newFakeChainAddressDeriver()
	catalog := newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
		newAllocationPolicy(
			"bitcoin-mainnet-native-segwit",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			"xpub-main",
			"fingerprint-main-native-segwit",
			"m/84'/0'/0'",
		),
	})
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	response, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 120000,
		CustomerReference:   " order-duplicate ",
		IdempotencyKey:      " idem-duplicate ",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if idempotencyStore.findCalls != 1 {
		t.Fatalf("expected duplicate idempotency lookup call count 1, got %d", idempotencyStore.findCalls)
	}
	if idempotencyStore.lastFindByKey.IdempotencyKey != "idem-duplicate" {
		t.Fatalf("unexpected idempotency key lookup input: got %q", idempotencyStore.lastFindByKey.IdempotencyKey)
	}
	if allocator.findIssuedByIDCalls != 1 {
		t.Fatalf("expected allocation lookup by id call count 1, got %d", allocator.findIssuedByIDCalls)
	}
	if allocator.lastFindIssuedByID.PaymentAddressID != 71 {
		t.Fatalf("unexpected payment address id lookup input: got %d", allocator.lastFindIssuedByID.PaymentAddressID)
	}
	if txManager.calls != 1 {
		t.Fatalf("expected one replay lookup transaction, got %d", txManager.calls)
	}
	if allocator.reserveFreshCalls != 0 {
		t.Fatalf("expected no fresh reservation for duplicate replay, got %d", allocator.reserveFreshCalls)
	}
	if deriver.calls != 0 {
		t.Fatalf("expected deriver not to be called, got %d", deriver.calls)
	}
	if response.PaymentAddressID != "71" {
		t.Fatalf("unexpected payment address id: got %q", response.PaymentAddressID)
	}
	if response.Address != "bc1qexistingduplicate" {
		t.Fatalf("unexpected address: got %q", response.Address)
	}
	if !response.IdempotencyReplayed {
		t.Fatal("expected duplicate replay response to be marked as replayed")
	}
}

func TestAllocatePaymentAddressUseCaseRejectsConflictingDuplicateIdempotencyKey(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{}
	txManager := newFakeUnitOfWork(allocator)
	idempotencyStore := &fakePaymentAddressIdempotencyStore{
		found: true,
		record: outport.PaymentAddressIdempotencyRecord{
			Chain:               valueobjects.SupportedChainBitcoin,
			IdempotencyKey:      "idem-conflict",
			AddressPolicyID:     "bitcoin-mainnet-native-segwit",
			ExpectedAmountMinor: 120000,
			CustomerReference:   "order-conflict",
			PaymentAddressID:    72,
		},
	}
	txManager.idempotencyStore = idempotencyStore
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		newFakeChainAddressDeriver(),
		newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
			newAllocationPolicy(
				"bitcoin-mainnet-native-segwit",
				valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
				string(valueobjects.BitcoinAddressSchemeNativeSegwit),
				"xpub-main",
				"fingerprint-main-native-segwit",
				"m/84'/0'/0'",
			),
		}),
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 99999,
		CustomerReference:   "order-conflict",
		IdempotencyKey:      "idem-conflict",
	})
	if !errors.Is(err, inport.ErrIdempotencyKeyConflict) {
		t.Fatalf("expected idempotency key conflict error, got %v", err)
	}
	if txManager.calls != 1 {
		t.Fatalf("expected one replay lookup transaction, got %d", txManager.calls)
	}
	if allocator.findIssuedByIDCalls != 0 {
		t.Fatalf("expected no allocation lookup on conflicting replay, got %d", allocator.findIssuedByIDCalls)
	}
	if allocator.reserveFreshCalls != 0 {
		t.Fatalf("expected no fresh reservation for conflicting replay, got %d", allocator.reserveFreshCalls)
	}
}

func TestAllocatePaymentAddressUseCaseResolvesConcurrentDuplicateAfterUniqueConflict(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{
		issuedByIDFound: true,
		issuedByID: entities.PaymentAddressAllocation{
			PaymentAddressID:    73,
			AddressPolicyID:     "bitcoin-mainnet-native-segwit",
			DerivationIndex:     12,
			ExpectedAmountMinor: 88000,
			CustomerReference:   "order-race",
			Status:              valueobjects.PaymentAddressAllocationStatusIssued,
			Chain:               valueobjects.SupportedChainBitcoin,
			Network:             valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:              string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			Address:             "bc1qracewinner",
			DerivationPath:      "m/84'/0'/0'/0/12",
		},
	}
	idempotencyStore := &fakePaymentAddressIdempotencyStore{
		findByKeyResults: []fakeFindPaymentAddressIdempotencyResult{
			{},
			{
				record: outport.PaymentAddressIdempotencyRecord{
					Chain:               valueobjects.SupportedChainBitcoin,
					IdempotencyKey:      "idem-race",
					AddressPolicyID:     "bitcoin-mainnet-native-segwit",
					ExpectedAmountMinor: 88000,
					CustomerReference:   "order-race",
					PaymentAddressID:    73,
				},
				found: true,
			},
		},
		claimErr: outport.ErrPaymentAddressIdempotencyKeyExists,
	}
	txManager := newFakeUnitOfWork(allocator)
	txManager.idempotencyStore = idempotencyStore
	deriver := newFakeChainAddressDeriver()
	deriver.output = newAllocateDeriveOutput("bc1qloserrace", "0/12")
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		deriver,
		newInMemoryAddressPolicyReader([]entities.AddressIssuancePolicy{
			newAllocationPolicy(
				"bitcoin-mainnet-native-segwit",
				valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
				string(valueobjects.BitcoinAddressSchemeNativeSegwit),
				"xpub-main",
				"fingerprint-main-native-segwit",
				"m/84'/0'/0'",
			),
		}),
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	response, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-native-segwit",
		ExpectedAmountMinor: 88000,
		CustomerReference:   "order-race",
		IdempotencyKey:      "idem-race",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if idempotencyStore.findCalls != 2 {
		t.Fatalf("expected duplicate idempotency lookup call count 2, got %d", idempotencyStore.findCalls)
	}
	if txManager.calls != 3 {
		t.Fatalf("expected three transaction attempts, got %d", txManager.calls)
	}
	if idempotencyStore.claimCalls != 1 {
		t.Fatalf("expected one idempotency claim attempt, got %d", idempotencyStore.claimCalls)
	}
	if allocator.reserveFreshCalls != 0 {
		t.Fatalf("expected no allocation reservation after duplicate claim, got %d", allocator.reserveFreshCalls)
	}
	if allocator.completeCalls != 0 {
		t.Fatalf("expected no allocation complete after duplicate claim, got %d", allocator.completeCalls)
	}
	trackingStore, ok := txManager.receiptTrackingStore.(*fakeAllocatePaymentReceiptTrackingStore)
	if !ok {
		t.Fatal("expected fake receipt tracking store")
	}
	if trackingStore.createCalls != 0 {
		t.Fatalf("expected no receipt tracking create after unique conflict, got %d", trackingStore.createCalls)
	}
	if response.PaymentAddressID != "73" {
		t.Fatalf("unexpected payment address id: got %q", response.PaymentAddressID)
	}
	if response.Address != "bc1qracewinner" {
		t.Fatalf("unexpected address: got %q", response.Address)
	}
	if deriver.calls != 0 {
		t.Fatalf("expected deriver not to be called after duplicate claim, got %d", deriver.calls)
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
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

	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(
			map[valueobjects.NetworkID]int32{
				valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet):  6,
				valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4): 2,
			},
			nil,
		),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
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

	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(
			nil,
			map[valueobjects.NetworkID]time.Duration{
				valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet):  48 * time.Hour,
				valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4): 24 * time.Hour,
			},
		),
		clock,
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			"xpub-main",
			"fingerprint-main-native-segwit",
			"m/84'/0'/0'",
		),
	})
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	response, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
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
		err:              expectedErr,
		allocationStore:  allocator,
		idempotencyStore: &fakePaymentAddressIdempotencyStore{},
	}
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		newFakeChainAddressDeriver(),
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			"xpub-main",
			"fingerprint-main-native-segwit",
			"m/84'/0'/0'",
		),
	})
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		deriver,
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
	txManager := newFakeUnitOfWork(allocator)
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		newFakeChainAddressDeriver(),
		newInMemoryAddressPolicyReader(nil),
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChain("eth"),
		AddressPolicyID:     "eth-mainnet",
		ExpectedAmountMinor: 1,
	})
	if !errors.Is(err, inport.ErrChainNotSupported) {
		t.Fatalf("expected chain not supported error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseRejectUnknownPolicy(t *testing.T) {
	allocator := &fakePaymentAddressAllocationStore{}
	txManager := newFakeUnitOfWork(allocator)
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		newFakeChainAddressDeriver(),
		newInMemoryAddressPolicyReader(nil),
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			"satoshi",
			8,
			"",
			"",
			"",
			"",
		),
	})
	allocator := &fakePaymentAddressAllocationStore{}
	txManager := newFakeUnitOfWork(allocator)
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		newFakeChainAddressDeriver(),
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	allocator := &fakePaymentAddressAllocationStore{reserveFreshErr: outport.ErrAddressIndexExhausted}
	txManager := newFakeUnitOfWork(allocator)
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		newFakeChainAddressDeriver(),
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	txManager := newFakeUnitOfWork(allocator)
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		&fakeChainAddressDeriver{
			supportedChains: map[valueobjects.SupportedChain]bool{
				valueobjects.SupportedChainBitcoin: true,
			},
			err: expectedErr,
		},
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 1,
		IdempotencyKey:      "idem-derive-failed",
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
	if allocator.lastFailedInput.Status != valueobjects.PaymentAddressAllocationStatusDerivationFailed {
		t.Fatalf("unexpected failed status: got %q", allocator.lastFailedInput.Status)
	}
	if allocator.completeCalls != 0 {
		t.Fatalf("expected complete allocation not to be called on derivation error")
	}
	idempotencyStore, ok := txManager.idempotencyStore.(*fakePaymentAddressIdempotencyStore)
	if !ok {
		t.Fatal("expected fake idempotency store")
	}
	if idempotencyStore.claimCalls != 1 {
		t.Fatalf("expected idempotency claim call count 1, got %d", idempotencyStore.claimCalls)
	}
	if idempotencyStore.releaseCalls != 1 {
		t.Fatalf("expected idempotency release call count 1, got %d", idempotencyStore.releaseCalls)
	}
	if idempotencyStore.completeCalls != 0 {
		t.Fatalf("expected idempotency complete not to be called, got %d", idempotencyStore.completeCalls)
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	txManager := newFakeUnitOfWork(allocator)
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		&fakeChainAddressDeriver{
			supportedChains: map[valueobjects.SupportedChain]bool{
				valueobjects.SupportedChainBitcoin: true,
			},
			err: expectedErr,
		},
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			"xpub-main",
			"fingerprint-main-legacy",
			"m/44'/0'/0'",
		),
	})
	allocator := &fakePaymentAddressAllocationStore{}
	txManager := newFakeUnitOfWork(allocator)
	useCase := newAllocatePaymentAddressUseCaseForTest(
		txManager,
		newFakeChainAddressDeriver(),
		catalog,
		policies.NewPaymentAddressAllocationIssuancePolicy(nil, nil),
		newAllocatePaymentAddressClock(),
	)

	_, err := useCase.Execute(context.Background(), dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
		AddressPolicyID:     "bitcoin-mainnet-legacy",
		ExpectedAmountMinor: 0,
	})
	if !errors.Is(err, inport.ErrInvalidExpectedAmount) {
		t.Fatalf("expected invalid expected amount error, got %v", err)
	}
}

func TestAllocatePaymentAddressUseCaseValidationMissingDependencies(t *testing.T) {
	input := dto.AllocatePaymentAddressInput{
		Chain:               valueobjects.SupportedChainBitcoin,
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
