package use_cases

import (
	"context"
	"errors"
	"strings"
	"time"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

const testXPubFingerprintAlgo = "sha256-trunc64-hex-v1"

type inMemoryAddressPolicyReader struct {
	ordered []entities.AddressPolicy
	byID    map[string]entities.AddressPolicy
}

var _ outport.AddressPolicyReader = (*inMemoryAddressPolicyReader)(nil)

func newInMemoryAddressPolicyReader(policies []entities.AddressPolicy) *inMemoryAddressPolicyReader {
	ordered := make([]entities.AddressPolicy, 0, len(policies))
	byID := make(map[string]entities.AddressPolicy, len(policies))

	for _, policy := range policies {
		normalized := policy.Normalize()
		if normalized.AddressPolicyID == "" {
			continue
		}
		if _, exists := byID[normalized.AddressPolicyID]; exists {
			continue
		}
		ordered = append(ordered, normalized)
		byID[normalized.AddressPolicyID] = normalized
	}

	return &inMemoryAddressPolicyReader{ordered: ordered, byID: byID}
}

func (r *inMemoryAddressPolicyReader) ListByChain(
	_ context.Context,
	chain value_objects.Chain,
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

func (r *inMemoryAddressPolicyReader) FindByID(
	_ context.Context,
	addressPolicyID string,
) (entities.AddressPolicy, bool, error) {
	policy, ok := r.byID[strings.TrimSpace(addressPolicyID)]
	if !ok {
		return entities.AddressPolicy{}, false, nil
	}
	return policy, true, nil
}

type fakePolicyBitcoinAddressDeriver struct {
	address             string
	err                 error
	derivationPath      string
	derivationPathErr   error
	lastNetwork         value_objects.BitcoinNetwork
	lastScheme          value_objects.BitcoinAddressScheme
	lastXPub            string
	lastIndex           uint32
	lastDerivationXPub  string
	lastDerivationIndex uint32
	calls               int
}

func (f *fakePolicyBitcoinAddressDeriver) DeriveAddress(
	network value_objects.BitcoinNetwork,
	scheme value_objects.BitcoinAddressScheme,
	xpub string,
	index uint32,
) (string, error) {
	f.calls++
	f.lastNetwork = network
	f.lastScheme = scheme
	f.lastXPub = xpub
	f.lastIndex = index
	if f.err != nil {
		return "", f.err
	}
	return f.address, nil
}

func (f *fakePolicyBitcoinAddressDeriver) DerivationPath(xpub string, index uint32) (string, error) {
	f.lastDerivationXPub = xpub
	f.lastDerivationIndex = index
	if f.derivationPathErr != nil {
		return "", f.derivationPathErr
	}
	if f.derivationPath == "" {
		return "0/0", nil
	}
	return f.derivationPath, nil
}

type fakePaymentAddressAllocationRepository struct {
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
	lastFailedInput       entities.PaymentAddressAllocation
	reopenCalls           int
	reserveFreshCalls     int
	completeCalls         int
	markFailedCalls       int
}

func (f *fakePaymentAddressAllocationRepository) ReopenFailedReservation(
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

func (f *fakePaymentAddressAllocationRepository) ReserveFresh(
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

func (f *fakePaymentAddressAllocationRepository) Complete(
	_ context.Context,
	input entities.PaymentAddressAllocation,
) error {
	f.completeCalls++
	f.lastCompleteInput = input
	return f.completeErr
}

func (f *fakePaymentAddressAllocationRepository) MarkDerivationFailed(
	_ context.Context,
	input entities.PaymentAddressAllocation,
) error {
	f.markFailedCalls++
	f.lastFailedInput = input
	return f.markFailedErr
}

type fakeUnitOfWork struct {
	err                       error
	calls                     int
	allocationRepository      outport.PaymentAddressAllocationRepository
	receiptTrackingRepository outport.PaymentReceiptTrackingRepository
	notificationRepository    outport.PaymentReceiptStatusNotificationRepository
}

func newFakeUnitOfWork(repository outport.PaymentAddressAllocationRepository) *fakeUnitOfWork {
	return &fakeUnitOfWork{
		allocationRepository:      repository,
		receiptTrackingRepository: &fakeAllocatePaymentReceiptTrackingRepository{},
		notificationRepository:    &fakeAllocatePaymentReceiptStatusNotificationRepository{},
	}
}

func (f *fakeUnitOfWork) WithinTransaction(
	_ context.Context,
	fn func(txRepositories outport.TxRepositories) error,
) error {
	f.calls++
	if f.err != nil {
		return f.err
	}
	if f.allocationRepository == nil {
		return errors.New("payment address allocation repository is not configured")
	}
	if f.receiptTrackingRepository == nil {
		return errors.New("payment receipt tracking repository is not configured")
	}
	return fn(outport.TxRepositories{
		PaymentAddressAllocation:         f.allocationRepository,
		PaymentReceiptTracking:           f.receiptTrackingRepository,
		PaymentReceiptStatusNotification: f.notificationRepository,
	})
}

type fakeAllocatePaymentReceiptTrackingRepository struct {
	registerErr                  error
	registerCalls                int
	lastRegisterPaymentAddressID int64
	lastRegisterConfirmations    int32
	lastRegisterExpiresAt        time.Time
}

func (f *fakeAllocatePaymentReceiptTrackingRepository) RegisterIssuedAllocation(
	_ context.Context,
	paymentAddressID int64,
	defaultRequiredConfirmations int32,
	expiresAt time.Time,
) (bool, error) {
	f.registerCalls++
	f.lastRegisterPaymentAddressID = paymentAddressID
	f.lastRegisterConfirmations = defaultRequiredConfirmations
	f.lastRegisterExpiresAt = expiresAt
	if f.registerErr != nil {
		return false, f.registerErr
	}
	return true, nil
}

func (f *fakeAllocatePaymentReceiptTrackingRepository) ClaimDue(
	_ context.Context,
	_ outport.ClaimPaymentReceiptTrackingsInput,
) ([]entities.PaymentReceiptTracking, error) {
	return nil, nil
}

func (f *fakeAllocatePaymentReceiptTrackingRepository) SaveObservation(
	_ context.Context,
	_ entities.PaymentReceiptTracking,
	_ time.Time,
	_ time.Time,
) error {
	return nil
}

func (f *fakeAllocatePaymentReceiptTrackingRepository) SavePollingError(
	_ context.Context,
	_ int64,
	_ string,
	_ time.Time,
	_ time.Time,
) error {
	return nil
}

type fakeAllocatePaymentReceiptStatusNotificationRepository struct{}

func (f *fakeAllocatePaymentReceiptStatusNotificationRepository) EnqueueStatusChanged(
	_ context.Context,
	_ outport.EnqueuePaymentReceiptStatusChangedInput,
) error {
	return nil
}

func (f *fakeAllocatePaymentReceiptStatusNotificationRepository) ClaimPending(
	_ context.Context,
	_ outport.ClaimPaymentReceiptStatusNotificationsInput,
) ([]entities.PaymentReceiptStatusNotification, error) {
	return nil, nil
}

func (f *fakeAllocatePaymentReceiptStatusNotificationRepository) MarkSent(
	_ context.Context,
	_ int64,
	_ time.Time,
) error {
	return nil
}

func (f *fakeAllocatePaymentReceiptStatusNotificationRepository) MarkRetryScheduled(
	_ context.Context,
	_ outport.MarkPaymentReceiptStatusNotificationRetryInput,
) error {
	return nil
}

func (f *fakeAllocatePaymentReceiptStatusNotificationRepository) MarkFailed(
	_ context.Context,
	_ outport.MarkPaymentReceiptStatusNotificationFailureInput,
) error {
	return nil
}
