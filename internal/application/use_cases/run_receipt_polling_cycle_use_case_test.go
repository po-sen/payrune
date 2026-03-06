package use_cases

import (
	"context"
	"errors"
	"testing"
	"time"

	"payrune/internal/application/dto"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

type fakeReceiptPollingClock struct {
	now time.Time
}

func (f *fakeReceiptPollingClock) NowUTC() time.Time {
	return f.now
}

type fakePaymentReceiptTrackingRepository struct {
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

func (f *fakePaymentReceiptTrackingRepository) RegisterIssuedAllocation(
	_ context.Context,
	_ int64,
	_ int32,
	_ time.Time,
) (bool, error) {
	return true, nil
}

func (f *fakePaymentReceiptTrackingRepository) ClaimDue(
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

func (f *fakePaymentReceiptTrackingRepository) SaveObservation(
	_ context.Context,
	tracking entities.PaymentReceiptTracking,
	now time.Time,
	nextPollAt time.Time,
) error {
	f.savedObservationTrackings = append(f.savedObservationTrackings, tracking)
	f.savedObservationPolledAtValues = append(f.savedObservationPolledAtValues, now)
	f.savedObservationNextPollAtValues = append(f.savedObservationNextPollAtValues, nextPollAt)
	return f.saveErr
}

func (f *fakePaymentReceiptTrackingRepository) SavePollingError(
	_ context.Context,
	paymentAddressID int64,
	errorReason string,
	now time.Time,
	nextPollAt time.Time,
) error {
	f.savedPollingErrorPaymentAddressID = append(f.savedPollingErrorPaymentAddressID, paymentAddressID)
	f.savedPollingErrorReasons = append(f.savedPollingErrorReasons, errorReason)
	f.savedPollingErrorPolledAt = append(f.savedPollingErrorPolledAt, now)
	f.savedPollingErrorNextPollAt = append(f.savedPollingErrorNextPollAt, nextPollAt)
	return f.savePollErr
}

type fakeBlockchainReceiptObserver struct {
	outputsByAddress map[string]outport.ObservePaymentAddressOutput
	errorsByAddress  map[string]error
	lastInputs       []outport.ObserveChainPaymentAddressInput
}

type fakeReceiptPollingUnitOfWork struct {
	repository outport.PaymentReceiptTrackingRepository
	err        error
	calls      int
}

func (f *fakeReceiptPollingUnitOfWork) WithinTransaction(
	_ context.Context,
	fn func(txRepositories outport.TxRepositories) error,
) error {
	f.calls++
	if f.err != nil {
		return f.err
	}
	return fn(outport.TxRepositories{
		PaymentReceiptTracking: f.repository,
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

	repository := &fakePaymentReceiptTrackingRepository{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{repository: repository}
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
		RunReceiptPollingCycleUseCaseConfig{},
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
	if output.FailedCount != 0 {
		t.Fatalf("unexpected failed count: got %d", output.FailedCount)
	}
	if repository.lastClaimInput.Limit != 10 {
		t.Fatalf("unexpected claim limit: got %d", repository.lastClaimInput.Limit)
	}
	if got := len(repository.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	if repository.savedObservationTrackings[0].Status != value_objects.PaymentReceiptStatusPaidConfirmed {
		t.Fatalf("unexpected saved status: got %q", repository.savedObservationTrackings[0].Status)
	}
	if got := repository.savedObservationNextPollAtValues[0]; !got.Equal(now.Add(20 * time.Second)) {
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

	repository := &fakePaymentReceiptTrackingRepository{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{repository: repository}
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
		RunReceiptPollingCycleUseCaseConfig{},
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
	if output.FailedCount != 1 {
		t.Fatalf("unexpected failed count: got %d", output.FailedCount)
	}
	if got := len(repository.savedPollingErrorPaymentAddressID); got != 1 {
		t.Fatalf("unexpected saved polling errors: got %d", got)
	}
	if repository.savedPollingErrorPaymentAddressID[0] != 202 {
		t.Fatalf("unexpected saved payment address id: got %d", repository.savedPollingErrorPaymentAddressID[0])
	}
	if repository.savedPollingErrorReasons[0] != "rpc timeout" {
		t.Fatalf("unexpected polling error reason: got %q", repository.savedPollingErrorReasons[0])
	}
	if unitOfWork.calls != 2 {
		t.Fatalf("unexpected uow calls: got %d, want 2", unitOfWork.calls)
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

	repository := &fakePaymentReceiptTrackingRepository{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{repository: repository}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{},
		errorsByAddress:  map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		RunReceiptPollingCycleUseCaseConfig{},
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
	if output.FailedCount != 1 {
		t.Fatalf("unexpected failed count: got %d", output.FailedCount)
	}
	if got := len(observer.lastInputs); got != 0 {
		t.Fatalf("unexpected observer calls: got %d", got)
	}
	if got := len(repository.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	if repository.savedObservationTrackings[0].Status != value_objects.PaymentReceiptStatusFailedExpired {
		t.Fatalf("unexpected saved status: got %q", repository.savedObservationTrackings[0].Status)
	}
	if repository.savedObservationTrackings[0].LastError != expiredReceiptReason {
		t.Fatalf("unexpected saved error: got %q", repository.savedObservationTrackings[0].LastError)
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

	repository := &fakePaymentReceiptTrackingRepository{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{repository: repository}
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
		RunReceiptPollingCycleUseCaseConfig{},
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
	if got := len(repository.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	updated := repository.savedObservationTrackings[0]
	if updated.Status != value_objects.PaymentReceiptStatusPaidUnconfirmed {
		t.Fatalf("unexpected status: got %q", updated.Status)
	}
	if updated.ExpiresAt == nil {
		t.Fatal("expected expires at to be set")
	}
	expectedExtendedAt := now.Add(defaultReceiptPaidUnconfirmedExpiryExtension)
	if !updated.ExpiresAt.Equal(expectedExtendedAt) {
		t.Fatalf("unexpected extended expires at: got %s, want %s", updated.ExpiresAt, expectedExtendedAt)
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

	repository := &fakePaymentReceiptTrackingRepository{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{repository: repository}
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
		RunReceiptPollingCycleUseCaseConfig{
			PaidUnconfirmedExpiryExtension: 6 * time.Hour,
		},
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
	if got := len(repository.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	updated := repository.savedObservationTrackings[0]
	if updated.ExpiresAt == nil {
		t.Fatal("expected expires at to be set")
	}
	expectedExtendedAt := now.Add(6 * time.Hour)
	if !updated.ExpiresAt.Equal(expectedExtendedAt) {
		t.Fatalf("unexpected extended expires at: got %s, want %s", updated.ExpiresAt, expectedExtendedAt)
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

	repository := &fakePaymentReceiptTrackingRepository{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{repository: repository}
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
		RunReceiptPollingCycleUseCaseConfig{
			PaidUnconfirmedExpiryExtension: 72 * time.Hour,
		},
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
	if got := len(repository.savedObservationTrackings); got != 1 {
		t.Fatalf("unexpected saved observations: got %d", got)
	}
	updated := repository.savedObservationTrackings[0]
	if updated.ExpiresAt == nil {
		t.Fatal("expected expires at to be set")
	}
	if !updated.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected expires at unchanged: got %s, want %s", updated.ExpiresAt, expiresAt)
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

	repository := &fakePaymentReceiptTrackingRepository{
		claimRows: []entities.PaymentReceiptTracking{tracking},
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{repository: repository}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{},
		errorsByAddress:  map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		RunReceiptPollingCycleUseCaseConfig{},
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
	if output.FailedCount != 1 {
		t.Fatalf("unexpected failed count: got %d", output.FailedCount)
	}
	if got := len(observer.lastInputs); got != 0 {
		t.Fatalf("unexpected observer calls: got %d", got)
	}
	if got := len(repository.savedPollingErrorReasons); got != 1 {
		t.Fatalf("unexpected polling errors: got %d", got)
	}
	if repository.savedPollingErrorReasons[0] != "issued at is required" {
		t.Fatalf("unexpected polling error reason: got %q", repository.savedPollingErrorReasons[0])
	}
}

func TestRunReceiptPollingCycleUseCaseExecuteValidation(t *testing.T) {
	repository := &fakePaymentReceiptTrackingRepository{}
	unitOfWork := &fakeReceiptPollingUnitOfWork{repository: repository}
	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		&fakeBlockchainReceiptObserver{
			outputsByAddress: map[string]outport.ObservePaymentAddressOutput{},
			errorsByAddress:  map[string]error{},
		},
		&fakeReceiptPollingClock{now: time.Now().UTC()},
		RunReceiptPollingCycleUseCaseConfig{},
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
	if repository.lastClaimInput.Chain != "eth" {
		t.Fatalf("unexpected normalized custom chain: got %q", repository.lastClaimInput.Chain)
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
	repository := &fakePaymentReceiptTrackingRepository{
		claimRows: []entities.PaymentReceiptTracking{},
	}
	unitOfWork := &fakeReceiptPollingUnitOfWork{repository: repository}
	observer := &fakeBlockchainReceiptObserver{
		outputsByAddress: map[string]outport.ObservePaymentAddressOutput{},
		errorsByAddress:  map[string]error{},
	}

	useCase := NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		observer,
		&fakeReceiptPollingClock{now: now},
		RunReceiptPollingCycleUseCaseConfig{},
	)

	_, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize: 10,
		Chain:     "bitcoin",
		Network:   "mainnet",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if repository.lastClaimInput.Chain != "bitcoin" {
		t.Fatalf("unexpected claim chain: got %q", repository.lastClaimInput.Chain)
	}
	if repository.lastClaimInput.Network != "mainnet" {
		t.Fatalf("unexpected claim network: got %q", repository.lastClaimInput.Network)
	}
}
