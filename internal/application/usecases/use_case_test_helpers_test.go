package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	applicationoutbox "payrune/internal/application/outbox"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

const testPublicKeyFingerprintAlgo = "sha256-trunc64-hex-v1"

type inMemoryAddressPolicyReader struct {
	ordered      []entities.AddressPolicy
	issuanceByID map[string]entities.AddressIssuancePolicy
}

var _ outport.AddressPolicyReader = (*inMemoryAddressPolicyReader)(nil)

func newAddressIssuancePolicy(
	addressPolicyID string,
	chain valueobjects.SupportedChain,
	network valueobjects.NetworkID,
	scheme string,
	minorUnit string,
	decimals uint8,
	accountPublicKey string,
	fingerprintAlgo string,
	fingerprint string,
	derivationPathPrefix string,
) entities.AddressIssuancePolicy {
	assetCode := "btc"
	assetType := "native"
	if chain == valueobjects.SupportedChainEthereum {
		assetCode = "eth"
		assetType = "native"
	}
	return entities.AddressIssuancePolicy{
		AddressPolicy: entities.AddressPolicy{
			AddressPolicyID: addressPolicyID,
			Chain:           chain,
			Network:         network,
			Scheme:          scheme,
			AssetCode:       assetCode,
			AssetType:       assetType,
			MinorUnit:       minorUnit,
			Decimals:        decimals,
		},
		DerivationConfig: valueobjects.AddressDerivationConfig{
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
	chain valueobjects.SupportedChain,
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
	supportedChains map[valueobjects.SupportedChain]bool
	output          outport.DeriveChainAddressOutput
	err             error
	lastInput       outport.DeriveChainAddressInput
	calls           int
}

func newFakeChainAddressDeriver() *fakeChainAddressDeriver {
	return &fakeChainAddressDeriver{
		supportedChains: map[valueobjects.SupportedChain]bool{
			valueobjects.SupportedChainBitcoin: true,
		},
		output: outport.DeriveChainAddressOutput{
			Address:                "bc1qdefault",
			RelativeDerivationPath: "0/0",
		},
	}
}

func (f *fakeChainAddressDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
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

type fakePaymentAddressStatusFinder struct {
	record    outport.PaymentAddressStatusRecord
	found     bool
	err       error
	lastInput outport.FindPaymentAddressStatusInput
	calls     int
}

func (f *fakePaymentAddressStatusFinder) FindByID(
	_ context.Context,
	input outport.FindPaymentAddressStatusInput,
) (outport.PaymentAddressStatusRecord, bool, error) {
	f.calls++
	f.lastInput = input
	if f.err != nil {
		return outport.PaymentAddressStatusRecord{}, false, f.err
	}
	if f.found {
		return f.record, true, nil
	}
	return outport.PaymentAddressStatusRecord{}, false, nil
}

type fakePaymentAddressAllocationStore struct {
	findIssuedByIDResults []fakeFindIssuedPaymentAddressAllocationResult
	issuedByID            entities.PaymentAddressAllocation
	issuedByIDFound       bool
	issuedByIDErr         error
	reopenReservation     entities.PaymentAddressAllocation
	reopenFound           bool
	reopenErr             error
	freshReservation      entities.PaymentAddressAllocation
	reserveFreshErr       error
	completeErr           error
	markFailedErr         error
	lastFindIssuedByID    outport.FindIssuedPaymentAddressAllocationByIDInput
	lastReopenInput       outport.ReservePaymentAddressAllocationInput
	lastReserveFreshInput outport.ReservePaymentAddressAllocationInput
	lastCompleteInput     entities.PaymentAddressAllocation
	lastCompleteIssuedAt  time.Time
	lastFailedInput       entities.PaymentAddressAllocation
	findIssuedByIDCalls   int
	reopenCalls           int
	reserveFreshCalls     int
	completeCalls         int
	markFailedCalls       int
}

type fakeFindIssuedPaymentAddressAllocationResult struct {
	allocation entities.PaymentAddressAllocation
	found      bool
	err        error
}

func (f *fakePaymentAddressAllocationStore) FindIssuedByID(
	_ context.Context,
	input outport.FindIssuedPaymentAddressAllocationByIDInput,
) (entities.PaymentAddressAllocation, bool, error) {
	f.findIssuedByIDCalls++
	f.lastFindIssuedByID = input
	if len(f.findIssuedByIDResults) >= f.findIssuedByIDCalls {
		result := f.findIssuedByIDResults[f.findIssuedByIDCalls-1]
		return result.allocation, result.found, result.err
	}
	if f.issuedByIDErr != nil {
		return entities.PaymentAddressAllocation{}, false, f.issuedByIDErr
	}
	if f.issuedByIDFound {
		return f.issuedByID, true, nil
	}
	return entities.PaymentAddressAllocation{}, false, nil
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
	evmFactoryStore      outport.EVMFactoryStore
	evmVaultStore        outport.EVMPaymentVaultStore
	allocationStore      outport.PaymentAddressAllocationStore
	idempotencyStore     outport.PaymentAddressIdempotencyStore
	receiptTrackingStore outport.PaymentReceiptTrackingStore
	notificationOutbox   outport.PaymentReceiptStatusNotificationOutbox
}

func newFakeUnitOfWork(store outport.PaymentAddressAllocationStore) *fakeUnitOfWork {
	return &fakeUnitOfWork{
		allocationStore:      store,
		idempotencyStore:     &fakePaymentAddressIdempotencyStore{},
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
	if f.evmFactoryStore == nil &&
		f.evmVaultStore == nil &&
		f.allocationStore == nil &&
		f.idempotencyStore == nil &&
		f.receiptTrackingStore == nil &&
		f.notificationOutbox == nil {
		return errors.New("transaction stores are not configured")
	}
	return fn(outport.TxScope{
		EVMFactoryRegistry:                     f.evmFactoryStore,
		EVMPaymentVaults:                       f.evmVaultStore,
		PaymentAddressAllocation:               f.allocationStore,
		PaymentAddressIdempotency:              f.idempotencyStore,
		PaymentReceiptTracking:                 f.receiptTrackingStore,
		PaymentReceiptStatusNotificationOutbox: f.notificationOutbox,
	})
}

type fakeEVMFactoryStore struct {
	record            outport.EVMFactoryRecord
	records           []outport.EVMFactoryRecord
	found             bool
	replaceErr        error
	listErr           error
	findErr           error
	replaceCalls      int
	listCalls         int
	findCalls         int
	lastReplaceInput  outport.ReplaceActiveEVMFactoryInput
	lastReplaceNow    time.Time
	lastFindByNetwork valueobjects.NetworkID
}

func (f *fakeEVMFactoryStore) ReplaceActive(
	_ context.Context,
	input outport.ReplaceActiveEVMFactoryInput,
	now time.Time,
) (outport.EVMFactoryRecord, error) {
	f.replaceCalls++
	f.lastReplaceInput = input
	f.lastReplaceNow = now
	if f.replaceErr != nil {
		return outport.EVMFactoryRecord{}, f.replaceErr
	}
	return f.record, nil
}

func (f *fakeEVMFactoryStore) ListActive(_ context.Context) ([]outport.EVMFactoryRecord, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.records, nil
}

func (f *fakeEVMFactoryStore) FindActiveByNetwork(
	_ context.Context,
	network valueobjects.NetworkID,
) (outport.EVMFactoryRecord, bool, error) {
	f.findCalls++
	f.lastFindByNetwork = network
	if f.findErr != nil {
		return outport.EVMFactoryRecord{}, false, f.findErr
	}
	if f.found {
		return f.record, true, nil
	}
	return outport.EVMFactoryRecord{}, false, nil
}

type fakePaymentAddressIdempotencyStore struct {
	findByKeyResults []fakeFindPaymentAddressIdempotencyResult
	record           outport.PaymentAddressIdempotencyRecord
	found            bool
	findErr          error
	claimErr         error
	releaseErr       error
	completeErr      error
	lastFindByKey    outport.FindPaymentAddressIdempotencyInput
	lastClaim        outport.ClaimPaymentAddressIdempotencyInput
	lastRelease      outport.ReleasePaymentAddressIdempotencyInput
	lastComplete     outport.CompletePaymentAddressIdempotencyInput
	findCalls        int
	claimCalls       int
	releaseCalls     int
	completeCalls    int
}

type fakeEVMPaymentVaultStore struct {
	record              outport.EVMPaymentVaultRecord
	err                 error
	createCalls         int
	lastCreateInput     outport.CreateEVMPaymentVaultInput
	sweepCandidates     []outport.EVMSweepCandidateRecord
	findSweepErr        error
	submittedSweepCalls int
	lastSubmittedSweep  outport.MarkEVMSweepSubmittedInput
	succeededSweepCalls int
	lastSucceededSweep  outport.MarkEVMSweepResultInput
	failedSweepCalls    int
	lastFailedSweep     outport.MarkEVMSweepResultInput
}

func (f *fakeEVMPaymentVaultStore) Create(
	_ context.Context,
	input outport.CreateEVMPaymentVaultInput,
) (outport.EVMPaymentVaultRecord, error) {
	f.createCalls++
	f.lastCreateInput = input
	if f.err != nil {
		return outport.EVMPaymentVaultRecord{}, f.err
	}
	if f.record.PaymentAddressID == 0 {
		f.record = outport.EVMPaymentVaultRecord{
			PaymentAddressID: input.PaymentAddressID,
			Network:          input.Network,
			FactoryID:        input.FactoryID,
			FactoryAddress:   input.FactoryAddress,
			CollectorAddress: input.CollectorAddress,
			TokenAddress:     input.TokenAddress,
			SaltHex:          input.SaltHex,
			PredictedAddress: input.PredictedAddress,
			DeployStatus:     "predicted",
			SweepStatus:      "pending",
		}
	}
	return f.record, nil
}

func (f *fakeEVMPaymentVaultStore) FindSweepCandidates(
	_ context.Context,
	_ outport.FindEVMSweepCandidatesInput,
) ([]outport.EVMSweepCandidateRecord, error) {
	if f.findSweepErr != nil {
		return nil, f.findSweepErr
	}
	records := make([]outport.EVMSweepCandidateRecord, len(f.sweepCandidates))
	copy(records, f.sweepCandidates)
	return records, nil
}

func (f *fakeEVMPaymentVaultStore) MarkSweepSubmitted(
	_ context.Context,
	input outport.MarkEVMSweepSubmittedInput,
) error {
	f.submittedSweepCalls++
	f.lastSubmittedSweep = input
	return nil
}

func (f *fakeEVMPaymentVaultStore) MarkSweepSucceeded(
	_ context.Context,
	input outport.MarkEVMSweepResultInput,
) error {
	f.succeededSweepCalls++
	f.lastSucceededSweep = input
	return nil
}

func (f *fakeEVMPaymentVaultStore) MarkSweepFailed(
	_ context.Context,
	input outport.MarkEVMSweepResultInput,
) error {
	f.failedSweepCalls++
	f.lastFailedSweep = input
	return nil
}

type fakeFindPaymentAddressIdempotencyResult struct {
	record outport.PaymentAddressIdempotencyRecord
	found  bool
	err    error
}

func (f *fakePaymentAddressIdempotencyStore) FindByKey(
	_ context.Context,
	input outport.FindPaymentAddressIdempotencyInput,
) (outport.PaymentAddressIdempotencyRecord, bool, error) {
	f.findCalls++
	f.lastFindByKey = input
	if len(f.findByKeyResults) >= f.findCalls {
		result := f.findByKeyResults[f.findCalls-1]
		return result.record, result.found, result.err
	}
	if f.findErr != nil {
		return outport.PaymentAddressIdempotencyRecord{}, false, f.findErr
	}
	if f.found {
		return f.record, true, nil
	}
	return outport.PaymentAddressIdempotencyRecord{}, false, nil
}

func (f *fakePaymentAddressIdempotencyStore) Claim(
	_ context.Context,
	input outport.ClaimPaymentAddressIdempotencyInput,
) (outport.PaymentAddressIdempotencyRecord, error) {
	f.claimCalls++
	f.lastClaim = input
	if f.claimErr != nil {
		return outport.PaymentAddressIdempotencyRecord{}, f.claimErr
	}
	return outport.PaymentAddressIdempotencyRecord{
		Chain:               input.Chain,
		IdempotencyKey:      input.IdempotencyKey,
		AddressPolicyID:     input.AddressPolicyID,
		ExpectedAmountMinor: input.ExpectedAmountMinor,
		CustomerReference:   input.CustomerReference,
	}, nil
}

func (f *fakePaymentAddressIdempotencyStore) Complete(
	_ context.Context,
	input outport.CompletePaymentAddressIdempotencyInput,
) error {
	f.completeCalls++
	f.lastComplete = input
	return f.completeErr
}

func (f *fakePaymentAddressIdempotencyStore) Release(
	_ context.Context,
	input outport.ReleasePaymentAddressIdempotencyInput,
) error {
	f.releaseCalls++
	f.lastRelease = input
	return f.releaseErr
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
