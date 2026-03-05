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
	registerCount int
	registerErr   error
	claimRows     []entities.PaymentReceiptTracking
	claimErr      error
	saveErr       error
	savePollErr   error

	lastRegisterNow                   time.Time
	lastRegisterDefaultConfirmations  int32
	lastRegisterChain                 string
	lastRegisterNetwork               string
	lastClaimInput                    outport.ClaimPaymentReceiptTrackingsInput
	savedObservationTrackings         []entities.PaymentReceiptTracking
	savedObservationNextPollAtValues  []time.Time
	savedObservationPolledAtValues    []time.Time
	savedPollingErrorPaymentAddressID []int64
	savedPollingErrorReasons          []string
	savedPollingErrorNextPollAt       []time.Time
	savedPollingErrorPolledAt         []time.Time
}

func (f *fakePaymentReceiptTrackingRepository) RegisterMissingIssued(
	_ context.Context,
	now time.Time,
	defaultRequiredConfirmations int32,
	chain string,
	network string,
) (int, error) {
	f.lastRegisterNow = now
	f.lastRegisterDefaultConfirmations = defaultRequiredConfirmations
	f.lastRegisterChain = chain
	f.lastRegisterNetwork = network
	if f.registerErr != nil {
		return 0, f.registerErr
	}
	return f.registerCount, nil
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
		registerCount: 2,
		claimRows:     []entities.PaymentReceiptTracking{tracking},
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
	)

	output, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize:                    10,
		PollInterval:                 20 * time.Second,
		ClaimTTL:                     9 * time.Second,
		DefaultRequiredConfirmations: 3,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if output.RegisteredCount != 2 {
		t.Fatalf("unexpected registered count: got %d", output.RegisteredCount)
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
	if repository.lastRegisterDefaultConfirmations != 3 {
		t.Fatalf("unexpected default confirmations: got %d", repository.lastRegisterDefaultConfirmations)
	}
	if repository.lastRegisterChain != "" {
		t.Fatalf("unexpected register chain: got %q", repository.lastRegisterChain)
	}
	if repository.lastRegisterNetwork != "" {
		t.Fatalf("unexpected register network: got %q", repository.lastRegisterNetwork)
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
	if got := repository.savedObservationNextPollAtValues[0]; !got.Equal(now.Add(24 * time.Hour)) {
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
		registerCount: 0,
		claimRows:     []entities.PaymentReceiptTracking{tracking},
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
	if repository.lastRegisterChain != "eth" {
		t.Fatalf("unexpected normalized custom chain: got %q", repository.lastRegisterChain)
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
		registerCount: 0,
		claimRows:     []entities.PaymentReceiptTracking{},
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
	)

	_, err := useCase.Execute(context.Background(), dto.RunReceiptPollingCycleInput{
		BatchSize: 10,
		Chain:     "bitcoin",
		Network:   "mainnet",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if repository.lastRegisterChain != "bitcoin" {
		t.Fatalf("unexpected register chain: got %q", repository.lastRegisterChain)
	}
	if repository.lastRegisterNetwork != "mainnet" {
		t.Fatalf("unexpected register network: got %q", repository.lastRegisterNetwork)
	}
	if repository.lastClaimInput.Chain != "bitcoin" {
		t.Fatalf("unexpected claim chain: got %q", repository.lastClaimInput.Chain)
	}
	if repository.lastClaimInput.Network != "mainnet" {
		t.Fatalf("unexpected claim network: got %q", repository.lastClaimInput.Network)
	}
}
