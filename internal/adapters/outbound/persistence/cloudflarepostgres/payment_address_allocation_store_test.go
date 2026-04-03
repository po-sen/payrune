package cloudflarepostgres

import (
	"context"
	"strings"
	"testing"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

type stubAllocationExecutor struct {
	execCalls         []stubExecCall
	queryRowCalls     []stubQueryRowCall
	queryRowResponses []row
}

type stubExecCall struct {
	query string
	args  []any
}

type stubQueryRowCall struct {
	query string
	args  []any
}

func (s *stubAllocationExecutor) ExecContext(_ context.Context, query string, args ...any) (result, error) {
	s.execCalls = append(s.execCalls, stubExecCall{query: query, args: args})
	return execResult{rowsAffected: 1}, nil
}

func (s *stubAllocationExecutor) QueryContext(_ context.Context, _ string, _ ...any) (rows, error) {
	return &sliceRows{}, nil
}

func (s *stubAllocationExecutor) QueryRowContext(_ context.Context, query string, args ...any) row {
	s.queryRowCalls = append(s.queryRowCalls, stubQueryRowCall{query: query, args: args})
	if len(s.queryRowResponses) == 0 {
		return errorRow{err: nil}
	}
	row := s.queryRowResponses[0]
	s.queryRowResponses = s.queryRowResponses[1:]
	return row
}

func newCloudflareAllocationStoreTestPolicy() policies.AddressIssuancePolicy {
	return policies.AddressIssuancePolicy{
		AddressPolicyID: "bitcoin-mainnet-native-segwit",
		Chain:           valueobjects.SupportedChainBitcoin,
		Network:         valueobjects.NetworkIDMainnet,
		Scheme:          valueobjects.AddressSchemeNativeSegwit,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef: "xpub-main",
		},
	}.Normalize()
}

func newCloudflareReservePaymentAddressAllocationInput(customerReference string) outport.ReservePaymentAddressAllocationInput {
	return outport.ReservePaymentAddressAllocationInput{
		IssuancePolicy:      newCloudflareAllocationStoreTestPolicy(),
		ExpectedAmountMinor: 125000,
		CustomerReference:   customerReference,
	}
}

func TestPaymentAddressAllocationStoreFindIssuedByIDSuccess(t *testing.T) {
	executor := &stubAllocationExecutor{
		queryRowResponses: []row{
			valueRow{values: []any{
				int64(199),
				"bitcoin-mainnet-native-segwit",
				int64(21),
				int64(125000),
				"order-lookup",
				"bitcoin",
				"mainnet",
				"nativeSegwit",
				"bc1qlookup",
				"",
			}},
		},
	}
	store := NewPaymentAddressAllocationStore(executor)

	allocation, found, err := store.FindIssuedByID(
		context.Background(),
		outport.FindIssuedPaymentAddressAllocationByIDInput{PaymentAddressID: 199},
	)
	if err != nil {
		t.Fatalf("FindIssuedByID returned error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if allocation.SlotIndex != 21 {
		t.Fatalf("unexpected slot index: got %d", allocation.SlotIndex)
	}
	if len(executor.queryRowCalls) != 1 {
		t.Fatalf("unexpected query row call count: got %d", len(executor.queryRowCalls))
	}
	if strings.Contains(executor.queryRowCalls[0].query, "address_policy_allocation_states") {
		t.Fatalf("expected single-table issued lookup, got %q", executor.queryRowCalls[0].query)
	}
}

func TestPaymentAddressAllocationStoreCompleteSuccess(t *testing.T) {
	executor := &stubAllocationExecutor{}
	store := NewPaymentAddressAllocationStore(executor)
	issuedAt := time.Date(2026, 3, 7, 9, 0, 0, 0, time.UTC)

	err := store.Complete(context.Background(), outport.CompletePaymentAddressAllocationInput{
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 44,
			AddressPolicyID:  "bitcoin-mainnet-native-segwit",
			Chain:            valueobjects.SupportedChainBitcoin,
			Network:          valueobjects.NetworkIDMainnet,
			Scheme:           valueobjects.AddressSchemeNativeSegwit,
			Address:          " bc1qallocated ",
		},
		SweepMaterialJSON: `{"material_type":"bitcoin_hd"}`,
		IssuedAt:          issuedAt,
	})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if len(executor.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(executor.execCalls))
	}
	if !strings.Contains(executor.execCalls[0].query, "UPDATE address_policy_allocations") {
		t.Fatalf("unexpected exec query: %q", executor.execCalls[0].query)
	}
	if len(executor.execCalls[0].args) != 7 {
		t.Fatalf("unexpected exec args: %+v", executor.execCalls[0].args)
	}
	if got := executor.execCalls[0].args[5]; got != `{"material_type":"bitcoin_hd"}` {
		t.Fatalf("unexpected sweep material json: %v", got)
	}
}

func TestPaymentAddressAllocationStoreCompleteRejectsInvalidSweepMaterialInput(t *testing.T) {
	executor := &stubAllocationExecutor{}
	store := NewPaymentAddressAllocationStore(executor)

	err := store.Complete(context.Background(), outport.CompletePaymentAddressAllocationInput{
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 44,
			AddressPolicyID:  "bitcoin-mainnet-native-segwit",
			Chain:            valueobjects.SupportedChainBitcoin,
			Network:          valueobjects.NetworkIDMainnet,
			Scheme:           valueobjects.AddressSchemeNativeSegwit,
			Address:          "bc1qallocated",
		},
		SweepMaterialJSON: " ",
		IssuedAt:          time.Date(2026, 3, 7, 9, 0, 0, 0, time.UTC),
	})
	if err != outport.ErrPaymentAddressAllocationStoreFailed {
		t.Fatalf("expected ErrPaymentAddressAllocationStoreFailed, got %v", err)
	}
	if len(executor.execCalls) != 0 {
		t.Fatalf("expected no exec calls, got %+v", executor.execCalls)
	}
}

func TestPaymentAddressAllocationStoreMarkDerivationFailedSuccess(t *testing.T) {
	executor := &stubAllocationExecutor{}
	store := NewPaymentAddressAllocationStore(executor)

	err := store.MarkDerivationFailed(context.Background(), entities.PaymentAddressAllocation{
		PaymentAddressID:        44,
		DerivationFailureReason: valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed,
	})
	if err != nil {
		t.Fatalf("MarkDerivationFailed returned error: %v", err)
	}
	if len(executor.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(executor.execCalls))
	}
	if !strings.Contains(executor.execCalls[0].query, "failure_reason = $2") {
		t.Fatalf("expected allocation-table failure reason update, got %q", executor.execCalls[0].query)
	}
}

func TestPaymentAddressAllocationStoreReopenFailedReservationUsesAllocationTable(t *testing.T) {
	executor := &stubAllocationExecutor{
		queryRowResponses: []row{
			valueRow{values: []any{int64(77), int64(15)}},
		},
	}
	store := NewPaymentAddressAllocationStore(executor)
	input := newCloudflareReservePaymentAddressAllocationInput(" order-5 ")

	allocation, reopened, err := store.ReopenFailedReservation(context.Background(), input)
	if err != nil {
		t.Fatalf("ReopenFailedReservation returned error: %v", err)
	}
	if !reopened {
		t.Fatal("expected reopened=true")
	}
	if allocation.SlotIndex != 15 {
		t.Fatalf("unexpected slot index: got %d", allocation.SlotIndex)
	}
	if len(executor.queryRowCalls) != 1 {
		t.Fatalf("expected 1 query row call, got %d", len(executor.queryRowCalls))
	}
	if strings.Contains(executor.queryRowCalls[0].query, "address_policy_allocation_states") {
		t.Fatalf("expected reopen query to stay on allocation table, got %q", executor.queryRowCalls[0].query)
	}
	if len(executor.queryRowCalls[0].args) != 2 ||
		executor.queryRowCalls[0].args[0] != input.IssuancePolicy.AddressPolicyID ||
		executor.queryRowCalls[0].args[1] != input.IssuancePolicy.IssuanceConfig.AddressSpaceRef {
		t.Fatalf("unexpected policy lookup args: %+v", executor.queryRowCalls[0].args)
	}
	if len(executor.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(executor.execCalls))
	}
}

func TestPaymentAddressAllocationStoreReserveFreshUsesAllocationTableAndPolicyCursor(t *testing.T) {
	executor := &stubAllocationExecutor{
		queryRowResponses: []row{
			valueRow{values: []any{int64(31)}},
			valueRow{values: []any{int64(155)}},
		},
	}
	store := NewPaymentAddressAllocationStore(executor)
	input := newCloudflareReservePaymentAddressAllocationInput(" order-6 ")

	allocation, err := store.ReserveFresh(context.Background(), input)
	if err != nil {
		t.Fatalf("ReserveFresh returned error: %v", err)
	}
	if allocation.PaymentAddressID != 155 {
		t.Fatalf("unexpected payment address id: got %d", allocation.PaymentAddressID)
	}
	if allocation.SlotIndex != 31 {
		t.Fatalf("unexpected slot index: got %d", allocation.SlotIndex)
	}
	if len(executor.execCalls) != 2 {
		t.Fatalf("expected 2 exec calls, got %d", len(executor.execCalls))
	}
	if !strings.Contains(executor.execCalls[0].query, "address_space_ref") {
		t.Fatalf("expected cursor seed to stay source-aware, got %q", executor.execCalls[0].query)
	}
	if len(executor.queryRowCalls) != 2 {
		t.Fatalf("expected 2 query row calls, got %d", len(executor.queryRowCalls))
	}
	if !strings.Contains(executor.queryRowCalls[1].query, "INSERT INTO address_policy_allocations") {
		t.Fatalf("expected allocation insert, got %q", executor.queryRowCalls[1].query)
	}
	if len(executor.queryRowCalls[1].args) != 5 {
		t.Fatalf("unexpected allocation insert args: %+v", executor.queryRowCalls[1].args)
	}
	if executor.queryRowCalls[1].args[1] != input.IssuancePolicy.IssuanceConfig.AddressSpaceRef {
		t.Fatalf("expected address space ref in allocation insert, got %+v", executor.queryRowCalls[1].args)
	}
	if executor.queryRowCalls[1].args[2] != int64(31) {
		t.Fatalf("expected slot index 31 in allocation insert, got %+v", executor.queryRowCalls[1].args)
	}
	if len(executor.execCalls[0].args) != 2 ||
		executor.execCalls[0].args[0] != input.IssuancePolicy.AddressPolicyID ||
		executor.execCalls[0].args[1] != input.IssuancePolicy.IssuanceConfig.AddressSpaceRef {
		t.Fatalf("unexpected cursor seed args: %+v", executor.execCalls[0].args)
	}
	if len(executor.queryRowCalls[0].args) != 2 ||
		executor.queryRowCalls[0].args[0] != input.IssuancePolicy.AddressPolicyID ||
		executor.queryRowCalls[0].args[1] != input.IssuancePolicy.IssuanceConfig.AddressSpaceRef {
		t.Fatalf("unexpected policy cursor lookup args: %+v", executor.queryRowCalls[0].args)
	}
	if len(executor.execCalls[1].args) != 2 ||
		executor.execCalls[1].args[0] != input.IssuancePolicy.AddressPolicyID ||
		executor.execCalls[1].args[1] != input.IssuancePolicy.IssuanceConfig.AddressSpaceRef {
		t.Fatalf("unexpected cursor update args: %+v", executor.execCalls[1].args)
	}
}
