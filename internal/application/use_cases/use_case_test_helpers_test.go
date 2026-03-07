package use_cases

import (
	"context"
	"errors"
	"strings"
	"time"

	applicationoutbox "payrune/internal/application/outbox"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/value_objects"
)

const testPublicKeyFingerprintAlgo = "sha256-trunc64-hex-v1"

type inMemoryAddressPolicyReader struct {
	ordered      []entities.AddressPolicy
	issuanceByID map[string]entities.AddressIssuancePolicy
}

var _ outport.AddressPolicyReader = (*inMemoryAddressPolicyReader)(nil)

func newAddressIssuancePolicy(
	addressPolicyID string,
	chain value_objects.SupportedChain,
	network value_objects.NetworkID,
	scheme string,
	minorUnit string,
	decimals uint8,
	accountPublicKey string,
	fingerprintAlgo string,
	fingerprint string,
	derivationPathPrefix string,
) entities.AddressIssuancePolicy {
	return entities.AddressIssuancePolicy{
		AddressPolicy: entities.AddressPolicy{
			AddressPolicyID: addressPolicyID,
			Chain:           chain,
			Network:         network,
			Scheme:          scheme,
			MinorUnit:       minorUnit,
			Decimals:        decimals,
		},
		DerivationConfig: value_objects.AddressDerivationConfig{
			AccountPublicKey:         accountPublicKey,
			PublicKeyFingerprintAlgo: fingerprintAlgo,
			PublicKeyFingerprint:     fingerprint,
			DerivationPathPrefix:     derivationPathPrefix,
		},
	}.Normalize()
}

func newInMemoryAddressPolicyReader(policies []entities.AddressIssuancePolicy) *inMemoryAddressPolicyReader {
	ordered := make([]entities.AddressPolicy, 0, len(policies))
	issuanceByID := make(map[string]entities.AddressIssuancePolicy, len(policies))

	for _, policy := range policies {
		normalized := policy.Normalize()
		if normalized.AddressPolicy.AddressPolicyID == "" {
			continue
		}
		if _, exists := issuanceByID[normalized.AddressPolicy.AddressPolicyID]; exists {
			continue
		}
		ordered = append(ordered, normalized.AddressPolicy)
		issuanceByID[normalized.AddressPolicy.AddressPolicyID] = normalized
	}

	return &inMemoryAddressPolicyReader{
		ordered:      ordered,
		issuanceByID: issuanceByID,
	}
}

func (r *inMemoryAddressPolicyReader) ListByChain(
	_ context.Context,
	chain value_objects.SupportedChain,
) ([]entities.AddressPolicy, error) {
	policies := make([]entities.AddressPolicy, 0)
	for _, policy := range r.ordered {
		if policy.Chain != chain {
			continue
		}
		policies = append(policies, policy)
	}
	return policies, nil
}

func (r *inMemoryAddressPolicyReader) FindIssuanceByID(
	_ context.Context,
	addressPolicyID string,
) (entities.AddressIssuancePolicy, bool, error) {
	policy, ok := r.issuanceByID[strings.TrimSpace(addressPolicyID)]
	if !ok {
		return entities.AddressIssuancePolicy{}, false, nil
	}
	return policy, true, nil
}

type fakeChainAddressDeriver struct {
	supportedChains map[value_objects.SupportedChain]bool
	output          outport.DeriveChainAddressOutput
	err             error
	lastInput       outport.DeriveChainAddressInput
	calls           int
}

func newFakeChainAddressDeriver() *fakeChainAddressDeriver {
	return &fakeChainAddressDeriver{
		supportedChains: map[value_objects.SupportedChain]bool{
			value_objects.SupportedChainBitcoin: true,
		},
		output: outport.DeriveChainAddressOutput{
			Address:                "bc1qdefault",
			RelativeDerivationPath: "0/0",
		},
	}
}

func (f *fakeChainAddressDeriver) SupportsChain(chain value_objects.SupportedChain) bool {
	return f.supportedChains[chain]
}

func (f *fakeChainAddressDeriver) DeriveAddress(
	_ context.Context,
	input outport.DeriveChainAddressInput,
) (outport.DeriveChainAddressOutput, error) {
	f.calls++
	f.lastInput = input
	if f.err != nil {
		return outport.DeriveChainAddressOutput{}, f.err
	}
	return f.output, nil
}

type fakePaymentAddressAllocationStore struct {
	reopenReservation     entities.PaymentAddressAllocation
	reopenFound           bool
	reopenErr             error
	freshReservation      entities.PaymentAddressAllocation
	reserveFreshErr       error
	completeErr           error
	markFailedErr         error
	lastReopenInput       outport.ReservePaymentAddressAllocationInput
	lastReserveFreshInput outport.ReservePaymentAddressAllocationInput
	lastCompleteInput     entities.PaymentAddressAllocation
	lastCompleteIssuedAt  time.Time
	lastFailedInput       entities.PaymentAddressAllocation
	reopenCalls           int
	reserveFreshCalls     int
	completeCalls         int
	markFailedCalls       int
}

func (f *fakePaymentAddressAllocationStore) ReopenFailedReservation(
	_ context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (entities.PaymentAddressAllocation, bool, error) {
	f.reopenCalls++
	f.lastReopenInput = input
	if f.reopenErr != nil {
		return entities.PaymentAddressAllocation{}, false, f.reopenErr
	}
	if f.reopenFound {
		return f.reopenReservation, true, nil
	}
	return entities.PaymentAddressAllocation{}, false, nil
}

func (f *fakePaymentAddressAllocationStore) ReserveFresh(
	_ context.Context,
	input outport.ReservePaymentAddressAllocationInput,
) (entities.PaymentAddressAllocation, error) {
	f.reserveFreshCalls++
	f.lastReserveFreshInput = input
	if f.reserveFreshErr != nil {
		return entities.PaymentAddressAllocation{}, f.reserveFreshErr
	}
	return f.freshReservation, nil
}

func (f *fakePaymentAddressAllocationStore) Complete(
	_ context.Context,
	input entities.PaymentAddressAllocation,
	issuedAt time.Time,
) error {
	f.completeCalls++
	f.lastCompleteInput = input
	f.lastCompleteIssuedAt = issuedAt
	return f.completeErr
}

func (f *fakePaymentAddressAllocationStore) MarkDerivationFailed(
	_ context.Context,
	input entities.PaymentAddressAllocation,
) error {
	f.markFailedCalls++
	f.lastFailedInput = input
	return f.markFailedErr
}

type fakeUnitOfWork struct {
	err                  error
	calls                int
	allocationStore      outport.PaymentAddressAllocationStore
	receiptTrackingStore outport.PaymentReceiptTrackingStore
	notificationOutbox   outport.PaymentReceiptStatusNotificationOutbox
}

func newFakeUnitOfWork(store outport.PaymentAddressAllocationStore) *fakeUnitOfWork {
	return &fakeUnitOfWork{
		allocationStore:      store,
		receiptTrackingStore: &fakeAllocatePaymentReceiptTrackingStore{},
		notificationOutbox:   &fakeAllocatePaymentReceiptStatusNotificationOutbox{},
	}
}

func (f *fakeUnitOfWork) WithinTransaction(
	_ context.Context,
	fn func(txScope outport.TxScope) error,
) error {
	f.calls++
	if f.err != nil {
		return f.err
	}
	if f.allocationStore == nil {
		return errors.New("payment address allocation store is not configured")
	}
	if f.receiptTrackingStore == nil {
		return errors.New("payment receipt tracking store is not configured")
	}
	return fn(outport.TxScope{
		PaymentAddressAllocation:               f.allocationStore,
		PaymentReceiptTracking:                 f.receiptTrackingStore,
		PaymentReceiptStatusNotificationOutbox: f.notificationOutbox,
	})
}

type fakeAllocatePaymentReceiptTrackingStore struct {
	createErr          error
	createCalls        int
	lastCreateTracking entities.PaymentReceiptTracking
	lastCreateNextPoll time.Time
}

func (f *fakeAllocatePaymentReceiptTrackingStore) Create(
	_ context.Context,
	tracking entities.PaymentReceiptTracking,
	nextPollAt time.Time,
) error {
	f.createCalls++
	f.lastCreateTracking = tracking
	f.lastCreateNextPoll = nextPollAt
	return f.createErr
}

func (f *fakeAllocatePaymentReceiptTrackingStore) ClaimDue(
	_ context.Context,
	_ outport.ClaimPaymentReceiptTrackingsInput,
) ([]entities.PaymentReceiptTracking, error) {
	return nil, nil
}

func (f *fakeAllocatePaymentReceiptTrackingStore) Save(
	_ context.Context,
	_ entities.PaymentReceiptTracking,
	_ time.Time,
	_ time.Time,
) error {
	return nil
}

type fakeAllocatePaymentReceiptStatusNotificationOutbox struct{}

func (f *fakeAllocatePaymentReceiptStatusNotificationOutbox) EnqueueStatusChanged(
	_ context.Context,
	_ events.PaymentReceiptStatusChanged,
) error {
	return nil
}

func (f *fakeAllocatePaymentReceiptStatusNotificationOutbox) ClaimPending(
	_ context.Context,
	_ outport.ClaimPaymentReceiptStatusNotificationsInput,
) ([]applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage, error) {
	return nil, nil
}

func (f *fakeAllocatePaymentReceiptStatusNotificationOutbox) SaveDeliveryResult(
	_ context.Context,
	_ policies.PaymentReceiptStatusNotificationDeliveryResult,
) error {
	return nil
}
