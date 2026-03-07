package use_cases

import (
	"context"
	"errors"
	"testing"
	"time"

	"payrune/internal/application/dto"
	applicationoutbox "payrune/internal/application/outbox"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/events"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/value_objects"
)

type fakeReceiptPollingClock struct {
	now time.Time
}

func (f *fakeReceiptPollingClock) NowUTC() time.Time {
	return f.now
}

type fakePaymentReceiptTrackingStore struct {
	claimRows   []entities.PaymentReceiptTracking
	claimErr    error
	saveErr     error
	savePollErr error

	lastClaimInput                    outport.ClaimPaymentReceiptTrackingsInput
	savedObservationTrackings         []entities.PaymentReceiptTracking
	savedObservationNextPollAtValues  []time.Time
	savedObservationPolledAtValues    []time.Time
	savedPollingErrorPaymentAddressID []int64
	savedPollingErrorReasons          []string
	savedPollingErrorNextPollAt       []time.Time
	savedPollingErrorPolledAt         []time.Time
}

func (f *fakePaymentReceiptTrackingStore) Create(
	_ context.Context,
	_ entities.PaymentReceiptTracking,
	_ time.Time,
) error {
	return nil
}

func (f *fakePaymentReceiptTrackingStore) ClaimDue(
	_ context.Context,
	input outport.ClaimPaymentReceiptTrackingsInput,
) ([]entities.PaymentReceiptTracking, error) {
	f.lastClaimInput = input
	if f.claimErr != nil {
		return nil, f.claimErr
	}
	rows := make([]entities.PaymentReceiptTracking, len(f.claimRows))
	copy(rows, f.claimRows)
	return rows, nil
}

func (f *fakePaymentReceiptTrackingStore) Save(
	_ context.Context,
	tracking entities.PaymentReceiptTracking,
	now time.Time,
	nextPollAt time.Time,
) error {
	f.savedObservationTrackings = append(f.savedObservationTrackings, tracking)
	f.savedObservationPolledAtValues = append(f.savedObservationPolledAtValues, now)
	f.savedObservationNextPollAtValues = append(f.savedObservationNextPollAtValues, nextPollAt)
	if tracking.LastError != "" {
		f.savedPollingErrorPaymentAddressID = append(f.savedPollingErrorPaymentAddressID, tracking.PaymentAddressID)
		f.savedPollingErrorReasons = append(f.savedPollingErrorReasons, tracking.LastError)
		f.savedPollingErrorPolledAt = append(f.savedPollingErrorPolledAt, now)
		f.savedPollingErrorNextPollAt = append(f.savedPollingErrorNextPollAt, nextPollAt)
		return f.savePollErr
	}
	return f.saveErr
}

type fakePaymentReceiptStatusNotificationOutbox struct {
	enqueueErr    error
	enqueueInputs []events.PaymentReceiptStatusChanged
}

func (f *fakePaymentReceiptStatusNotificationOutbox) EnqueueStatusChanged(
	_ context.Context,
	input events.PaymentReceiptStatusChanged,
) error {
	f.enqueueInputs = append(f.enqueueInputs, input)
	return f.enqueueErr
}

func (f *fakePaymentReceiptStatusNotificationOutbox) ClaimPending(
	_ context.Context,
	_ outport.ClaimPaymentReceiptStatusNotificationsInput,
) ([]applicationoutbox.PaymentReceiptStatusNotificationOutboxMessage, error) {
	return nil, nil
}

func (f *fakePaymentReceiptStatusNotificationOutbox) SaveDeliveryResult(
	_ context.Context,
	_ policies.PaymentReceiptStatusNotificationDeliveryResult,
) error {
	return nil
}

type fakeBlockchainReceiptObserver struct {
	outputsByAddress map[string]outport.ObservePaymentAddressOutput
	errorsByAddress  map[string]error
	lastInputs       []outport.ObserveChainPaymentAddressInput
}

type fakeReceiptPollingUnitOfWork struct {
	trackingStore      outport.PaymentReceiptTrackingStore
	notificationOutbox outport.PaymentReceiptStatusNotificationOutbox
	err                error
	calls              int
}

func (f *fakeReceiptPollingUnitOfWork) WithinTransaction(
	_ context.Context,
	fn func(txScope outport.TxScope) error,
) error {
	f.calls++
	if f.err != nil {
		return f.err
	}
	return fn(outport.TxScope{
		PaymentReceiptTracking:                 f.trackingStore,
		PaymentReceiptStatusNotificationOutbox: f.notificationOutbox,
	})
}

func (f *fakeBlockchainReceiptObserver) ObserveAddress(
	_ context.Context,
	input outport.ObserveChainPaymentAddressInput,
) (outport.ObservePaymentAddressOutput, error) {
	f.lastInputs = append(f.lastInputs, input)
	if err := f.errorsByAddress[input.Address]; err != nil {
		return outport.ObservePaymentAddressOutput{}, err
	}
	output, ok := f.outputsByAddress[input.Address]
	if !ok {
		return outport.ObservePaymentAddressOutput{}, nil
	}
	return output, nil
}

func TestRunReceiptPollingCycleUseCaseExecuteSuccess(t *testing.T) {
	now := time.Date(2026, 3, 5, 14, 0, 0, 0, time.UTC)
	tracking, err := entities.NewPaymentReceiptTracking(
		101,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qreceipt1",
		time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC),
		1000,
		2,
	)
	if err != nil {
		t.Fatalf("setup tracking: %v", err)
	}

	trackingStore := &fakePaymentReceiptTrackingStore{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	notificationOutbox := &fakePaymentReceiptStatusNotificationOutbox{}
	unitOfWork := &fakeReceiptPollingUnitOfWork{
		trackingStore:      trackingStore,
		notificationOutbox: notificationOutbox,
	}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{
			"tb1qreceipt1": {
				ObservedTotalMinor:    1200,
				ConfirmedTotalMinor:   1200,
				UnconfirmedTotalMinor: 0,
				ConflictTotalMinor:    0,
				LatestBlockHeight:     1000,
			},
		},
		errorsByAddress: map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(0),
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize:    10,
		PollInterval: 20 * time.Second,
		ClaimTTL:     9 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if output.ClaimedCount != 1 {
		t.Fatalf("unexpected claimed count: got %d", output.ClaimedCount)
	}
	if output.UpdatedCount != 1 {
		t.Fatalf("unexpected updated count: got %d", output.UpdatedCount)
	}
	if output.TerminalFailedCount != 0 {
		t.Fatalf("unexpected terminal failed count: got %d", output.TerminalFailedCount)
	}
	if output.ProcessingErrorCount != 0 {
		t.Fatalf("unexpected processing error count: got %d", output.ProcessingErrorCount)
	}
	if trackingStore.lastClaimInput.Limit != 10 {
		t.Fatalf("unexpected claim limit: got %d", trackingStore.lastClaimInput.Limit)
	}
	if got := len(trackingStore.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	if trackingStore.savedObservationTrackings[0].Status != value_objects.PaymentReceiptStatusPaidConfirmed {
		t.Fatalf("unexpected saved status: got %q", trackingStore.savedObservationTrackings[0].Status)
	}
	if got := trackingStore.savedObservationNextPollAtValues[0]; !got.Equal(now.Add(20 * time.Second)) {
		t.Fatalf("unexpected next poll at for confirmed status: got %s", got)
	}
	if unitOfWork.calls != 2 {
		t.Fatalf("unexpected uow calls: got %d, want 2", unitOfWork.calls)
	}
	if got := len(observer.lastInputs); got != 1 {
		t.Fatalf("unexpected observer call count: got %d", got)
	}
	if observer.lastInputs[0].IssuedAt.IsZero() {
		t.Fatal("expected issued at in observer input")
	}
	if got := len(notificationOutbox.enqueueInputs); got != 1 {
		t.Fatalf("unexpected enqueue count: got %d", got)
	}
	notification := notificationOutbox.enqueueInputs[0]
	if notification.PreviousStatus != value_objects.PaymentReceiptStatusWatching {
		t.Fatalf("unexpected previous status: got %q", notification.PreviousStatus)
	}
	if notification.CurrentStatus != value_objects.PaymentReceiptStatusPaidConfirmed {
		t.Fatalf("unexpected current status: got %q", notification.CurrentStatus)
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteObserverError(t *testing.T) {
	now := time.Date(2026, 3, 5, 14, 10, 0, 0, time.UTC)
	tracking, err := entities.NewPaymentReceiptTracking(
		202,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qreceipt2",
		time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC),
		500,
		1,
	)
	if err != nil {
		t.Fatalf("setup tracking: %v", err)
	}

	trackingStore := &fakePaymentReceiptTrackingStore{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	notificationOutbox := &fakePaymentReceiptStatusNotificationOutbox{}
	unitOfWork := &fakeReceiptPollingUnitOfWork{
		trackingStore:      trackingStore,
		notificationOutbox: notificationOutbox,
	}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{},
		errorsByAddress: map[string]error{
			"tb1qreceipt2": errors.New("rpc timeout"),
		},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(0),
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize:    10,
		PollInterval: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if output.UpdatedCount != 0 {
		t.Fatalf("unexpected updated count: got %d", output.UpdatedCount)
	}
	if output.TerminalFailedCount != 0 {
		t.Fatalf("unexpected terminal failed count: got %d", output.TerminalFailedCount)
	}
	if output.ProcessingErrorCount != 1 {
		t.Fatalf("unexpected processing error count: got %d", output.ProcessingErrorCount)
	}
	if got := len(trackingStore.savedPollingErrorPaymentAddressID); got != 1 {
		t.Fatalf("unexpected saved polling errors: got %d", got)
	}
	if trackingStore.savedPollingErrorPaymentAddressID[0] != 202 {
		t.Fatalf("unexpected saved payment address id: got %d", trackingStore.savedPollingErrorPaymentAddressID[0])
	}
	if trackingStore.savedPollingErrorReasons[0] != "rpc timeout" {
		t.Fatalf("unexpected polling error reason: got %q", trackingStore.savedPollingErrorReasons[0])
	}
	if unitOfWork.calls != 2 {
		t.Fatalf("unexpected uow calls: got %d, want 2", unitOfWork.calls)
	}
	if got := len(notificationOutbox.enqueueInputs); got != 0 {
		t.Fatalf("unexpected enqueue count: got %d", got)
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteReturnsErrorWhenEnqueueFails(t *testing.T) {
	now := time.Date(2026, 3, 5, 14, 12, 0, 0, time.UTC)
	tracking, err := entities.NewPaymentReceiptTracking(
		212,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qreceipt-enqueue-fail",
		time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC),
		500,
		1,
	)
	if err != nil {
		t.Fatalf("setup tracking: %v", err)
	}

	trackingStore := &fakePaymentReceiptTrackingStore{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	notificationOutbox := &fakePaymentReceiptStatusNotificationOutbox{
		enqueueErr: errors.New("enqueue failed"),
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{
		trackingStore:      trackingStore,
		notificationOutbox: notificationOutbox,
	}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{
			"tb1qreceipt-enqueue-fail": {
				ObservedTotalMinor:    500,
				ConfirmedTotalMinor:   500,
				UnconfirmedTotalMinor: 0,
				ConflictTotalMinor:    0,
				LatestBlockHeight:     1001,
			},
		},
		errorsByAddress: map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(0),
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize:    10,
		PollInterval: 30 * time.Second,
	})
	if err == nil {
		t.Fatal("expected enqueue error but got nil")
	}
	if output.UpdatedCount != 0 {
		t.Fatalf("unexpected updated count: got %d", output.UpdatedCount)
	}
	if got := len(notificationOutbox.enqueueInputs); got != 1 {
		t.Fatalf("unexpected enqueue count: got %d", got)
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteExpiredTracking(t *testing.T) {
	now := time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC)
	tracking, err := entities.NewPaymentReceiptTracking(
		303,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qexpired",
		time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC),
		500,
		1,
	)
	if err != nil {
		t.Fatalf("setup tracking: %v", err)
	}
	expiredAt := now.Add(-1 * time.Minute)
	tracking.ExpiresAt = &expiredAt

	trackingStore := &fakePaymentReceiptTrackingStore{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	notificationOutbox := &fakePaymentReceiptStatusNotificationOutbox{}
	unitOfWork := &fakeReceiptPollingUnitOfWork{
		trackingStore:      trackingStore,
		notificationOutbox: notificationOutbox,
	}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{},
		errorsByAddress:  map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(0),
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize:    10,
		PollInterval: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.ClaimedCount != 1 {
		t.Fatalf("unexpected claimed count: got %d", output.ClaimedCount)
	}
	if output.UpdatedCount != 0 {
		t.Fatalf("unexpected updated count: got %d", output.UpdatedCount)
	}
	if output.TerminalFailedCount != 1 {
		t.Fatalf("unexpected terminal failed count: got %d", output.TerminalFailedCount)
	}
	if output.ProcessingErrorCount != 0 {
		t.Fatalf("unexpected processing error count: got %d", output.ProcessingErrorCount)
	}
	if got := len(observer.lastInputs); got != 0 {
		t.Fatalf("unexpected observer calls: got %d", got)
	}
	if got := len(trackingStore.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	if trackingStore.savedObservationTrackings[0].Status != value_objects.PaymentReceiptStatusFailedExpired {
		t.Fatalf("unexpected saved status: got %q", trackingStore.savedObservationTrackings[0].Status)
	}
	if trackingStore.savedObservationTrackings[0].LastError != "payment window expired" {
		t.Fatalf("unexpected saved error: got %q", trackingStore.savedObservationTrackings[0].LastError)
	}
	if got := len(notificationOutbox.enqueueInputs); got != 1 {
		t.Fatalf("unexpected enqueue count: got %d", got)
	}
	if notificationOutbox.enqueueInputs[0].CurrentStatus != value_objects.PaymentReceiptStatusFailedExpired {
		t.Fatalf("unexpected current status: got %q", notificationOutbox.enqueueInputs[0].CurrentStatus)
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteExtendsExpiryOnTransitionToPaidUnconfirmed(t *testing.T) {
	now := time.Date(2026, 3, 6, 11, 0, 0, 0, time.UTC)
	tracking, err := entities.NewPaymentReceiptTracking(
		304,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qpaidunconfirmed",
		time.Date(2026, 3, 5, 11, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup tracking: %v", err)
	}
	expiresAt := now.Add(30 * time.Minute)
	tracking.ExpiresAt = &expiresAt

	trackingStore := &fakePaymentReceiptTrackingStore{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	notificationOutbox := &fakePaymentReceiptStatusNotificationOutbox{}
	unitOfWork := &fakeReceiptPollingUnitOfWork{
		trackingStore:      trackingStore,
		notificationOutbox: notificationOutbox,
	}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{
			"tb1qpaidunconfirmed": {
				ObservedTotalMinor:    1000,
				ConfirmedTotalMinor:   0,
				UnconfirmedTotalMinor: 1000,
				ConflictTotalMinor:    0,
				LatestBlockHeight:     101,
			},
		},
		errorsByAddress: map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(0),
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize:    10,
		PollInterval: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.UpdatedCount != 1 {
		t.Fatalf("unexpected updated count: got %d", output.UpdatedCount)
	}
	if got := len(trackingStore.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	updated := trackingStore.savedObservationTrackings[0]
	if updated.Status != value_objects.PaymentReceiptStatusPaidUnconfirmed {
		t.Fatalf("unexpected status: got %q", updated.Status)
	}
	if updated.ExpiresAt == nil {
		t.Fatal("expected expires at to be set")
	}
	expectedExtendedAt := now.Add(7 * 24 * time.Hour)
	if !updated.ExpiresAt.Equal(expectedExtendedAt) {
		t.Fatalf("unexpected extended expires at: got %s, want %s", updated.ExpiresAt, expectedExtendedAt)
	}
	if got := len(notificationOutbox.enqueueInputs); got != 1 {
		t.Fatalf("unexpected enqueue count: got %d", got)
	}
	if notificationOutbox.enqueueInputs[0].CurrentStatus != value_objects.PaymentReceiptStatusPaidUnconfirmed {
		t.Fatalf("unexpected current status: got %q", notificationOutbox.enqueueInputs[0].CurrentStatus)
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteUsesConfiguredPaidUnconfirmedExpiryExtension(t *testing.T) {
	now := time.Date(2026, 3, 6, 11, 30, 0, 0, time.UTC)
	tracking, err := entities.NewPaymentReceiptTracking(
		305,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qpaidunconfirmedcustom",
		time.Date(2026, 3, 5, 11, 30, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup tracking: %v", err)
	}
	expiresAt := now.Add(30 * time.Minute)
	tracking.ExpiresAt = &expiresAt

	trackingStore := &fakePaymentReceiptTrackingStore{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	notificationOutbox := &fakePaymentReceiptStatusNotificationOutbox{}
	unitOfWork := &fakeReceiptPollingUnitOfWork{
		trackingStore:      trackingStore,
		notificationOutbox: notificationOutbox,
	}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{
			"tb1qpaidunconfirmedcustom": {
				ObservedTotalMinor:    1000,
				ConfirmedTotalMinor:   0,
				UnconfirmedTotalMinor: 1000,
				ConflictTotalMinor:    0,
				LatestBlockHeight:     101,
			},
		},
		errorsByAddress: map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(6*time.Hour),
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize:    10,
		PollInterval: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.UpdatedCount != 1 {
		t.Fatalf("unexpected updated count: got %d", output.UpdatedCount)
	}
	if got := len(trackingStore.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	updated := trackingStore.savedObservationTrackings[0]
	if updated.ExpiresAt == nil {
		t.Fatal("expected expires at to be set")
	}
	expectedExtendedAt := now.Add(6 * time.Hour)
	if !updated.ExpiresAt.Equal(expectedExtendedAt) {
		t.Fatalf("unexpected extended expires at: got %s, want %s", updated.ExpiresAt, expectedExtendedAt)
	}
	if got := len(notificationOutbox.enqueueInputs); got != 1 {
		t.Fatalf("unexpected enqueue count: got %d", got)
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteDoesNotExtendWhenStatusUnchanged(t *testing.T) {
	now := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	tracking, err := entities.NewPaymentReceiptTracking(
		306,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qpaidunconfirmedsteady",
		time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("setup tracking: %v", err)
	}
	tracking.Status = value_objects.PaymentReceiptStatusPaidUnconfirmed
	expiresAt := now.Add(2 * time.Hour)
	tracking.ExpiresAt = &expiresAt

	trackingStore := &fakePaymentReceiptTrackingStore{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	notificationOutbox := &fakePaymentReceiptStatusNotificationOutbox{}
	unitOfWork := &fakeReceiptPollingUnitOfWork{
		trackingStore:      trackingStore,
		notificationOutbox: notificationOutbox,
	}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{
			"tb1qpaidunconfirmedsteady": {
				ObservedTotalMinor:    1000,
				ConfirmedTotalMinor:   0,
				UnconfirmedTotalMinor: 1000,
				ConflictTotalMinor:    0,
				LatestBlockHeight:     101,
			},
		},
		errorsByAddress: map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(72*time.Hour),
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize:    10,
		PollInterval: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.UpdatedCount != 1 {
		t.Fatalf("unexpected updated count: got %d", output.UpdatedCount)
	}
	if got := len(trackingStore.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	updated := trackingStore.savedObservationTrackings[0]
	if updated.ExpiresAt == nil {
		t.Fatal("expected expires at to be set")
	}
	if !updated.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected expires at unchanged: got %s, want %s", updated.ExpiresAt, expiresAt)
	}
	if got := len(notificationOutbox.enqueueInputs); got != 0 {
		t.Fatalf("unexpected enqueue count: got %d", got)
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteMissingIssuedAt(t *testing.T) {
	now := time.Date(2026, 3, 5, 14, 15, 0, 0, time.UTC)
	tracking, err := entities.NewPaymentReceiptTracking(
		203,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID("testnet4"),
		"tb1qreceipt3",
		time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC),
		500,
		1,
	)
	if err != nil {
		t.Fatalf("setup tracking: %v", err)
	}
	tracking.IssuedAt = time.Time{}

	trackingStore := &fakePaymentReceiptTrackingStore{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	notificationOutbox := &fakePaymentReceiptStatusNotificationOutbox{}
	unitOfWork := &fakeReceiptPollingUnitOfWork{
		trackingStore:      trackingStore,
		notificationOutbox: notificationOutbox,
	}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{},
		errorsByAddress:  map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(0),
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize:    10,
		PollInterval: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if output.UpdatedCount != 0 {
		t.Fatalf("unexpected updated count: got %d", output.UpdatedCount)
	}
	if output.TerminalFailedCount != 0 {
		t.Fatalf("unexpected terminal failed count: got %d", output.TerminalFailedCount)
	}
	if output.ProcessingErrorCount != 1 {
		t.Fatalf("unexpected processing error count: got %d", output.ProcessingErrorCount)
	}
	if got := len(observer.lastInputs); got != 0 {
		t.Fatalf("unexpected observer calls: got %d", got)
	}
	if got := len(trackingStore.savedPollingErrorReasons); got != 1 {
		t.Fatalf("unexpected polling errors: got %d", got)
	}
	if trackingStore.savedPollingErrorReasons[0] != "issued at is required" {
		t.Fatalf("unexpected polling error reason: got %q", trackingStore.savedPollingErrorReasons[0])
	}
	if got := len(notificationOutbox.enqueueInputs); got != 0 {
		t.Fatalf("unexpected enqueue count: got %d", got)
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteValidation(t *testing.T) {
	trackingStore := &fakePaymentReceiptTrackingStore{}
	unitOfWork := &fakeReceiptPollingUnitOfWork{trackingStore: trackingStore, notificationOutbox: &fakePaymentReceiptStatusNotificationOutbox{}}
	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		&fakeBlockchainReceiptObserver{
			outputsByAddress: map[string]outport.ObservePaymentAddressOutput{},
			errorsByAddress:  map[string]error{},
		},
		&fakeReceiptPollingClock{now: time.Now().UTC()},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(0),
	)

	_, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{BatchSize: 0})
	if err == nil {
		t.Fatal("expected validation error but got nil")
	}

	_, err = useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize: 10,
		Chain:     "eth",
	})
	if err != nil {
		t.Fatalf("expected nil error for custom chain scope, got %v", err)
	}
	if trackingStore.lastClaimInput.Chain != "eth" {
		t.Fatalf("unexpected normalized custom chain: got %q", trackingStore.lastClaimInput.Chain)
	}

	_, err = useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize: 10,
		Network:   "regtest",
	})
	if err == nil {
		t.Fatal("expected missing chain error but got nil")
	}

	_, err = useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize: 10,
		Chain:     "eth/mainnet",
	})
	if err == nil {
		t.Fatal("expected invalid chain error but got nil")
	}

	_, err = useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize: 10,
		Chain:     "eth",
		Network:   "main/net",
	})
	if err == nil {
		t.Fatal("expected invalid network error but got nil")
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteWithScope(t *testing.T) {
	now := time.Date(2026, 3, 5, 14, 20, 0, 0, time.UTC)
	trackingStore := &fakePaymentReceiptTrackingStore{
		claimRows: []entities.PaymentReceiptTracking{},
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{trackingStore: trackingStore, notificationOutbox: &fakePaymentReceiptStatusNotificationOutbox{}}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{},
		errorsByAddress:  map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		policies.NewPaymentReceiptTrackingLifecyclePolicy(0),
	)

	_, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize: 10,
		Chain:     "bitcoin",
		Network:   "mainnet",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if trackingStore.lastClaimInput.Chain != "bitcoin" {
		t.Fatalf("unexpected claim chain: got %q", trackingStore.lastClaimInput.Chain)
	}
	if trackingStore.lastClaimInput.Network != "mainnet" {
		t.Fatalf("unexpected claim network: got %q", trackingStore.lastClaimInput.Network)
	}
}
