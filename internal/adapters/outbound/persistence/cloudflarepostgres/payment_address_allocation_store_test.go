package cloudflarepostgres

import (
	"context"
	"strings"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
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

func newCloudflareAllocationStoreTestPolicy() entities.AddressIssuancePolicy {
	return entities.AddressIssuancePolicy{
		AddressPolicy: entities.AddressPolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
		},
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSourceRef: "xpub-main",
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

func TestPaymentAddressAllocationStoreReopenFailedReservationUsesAddressSourceRef(t *testing.T) {
	executor := &stubAllocationExecutor{
		queryRowResponses: []row{
			valueRow{values: []any{int64(99), int64(11)}},
		},
	}
	store := NewPaymentAddressAllocationStore(executor)
	input := newCloudflareReservePaymentAddressAllocationInput(" order-1 ")

	allocation, reopened, err := store.ReopenFailedReservation(context.Background(), input)
	if err != nil {
		t.Fatalf("ReopenFailedReservation returned error: %v", err)
	}
	if !reopened {
		t.Fatal("expected reopened=true")
	}
	if allocation.DerivationIndex != 11 {
		t.Fatalf("unexpected derivation index: got %d", allocation.DerivationIndex)
	}
	if len(executor.queryRowCalls) != 1 {
		t.Fatalf("unexpected query row call count: got %d", len(executor.queryRowCalls))
	}
	if !strings.Contains(executor.queryRowCalls[0].query, "address_source_ref = $2") {
		t.Fatalf("expected reopen query to filter by address_source_ref, got %q", executor.queryRowCalls[0].query)
	}
	if len(executor.queryRowCalls[0].args) != 2 || executor.queryRowCalls[0].args[1] != "xpub-main" {
		t.Fatalf("unexpected reopen args: %+v", executor.queryRowCalls[0].args)
	}
}

func TestPaymentAddressAllocationStoreReserveFreshUsesXPubOnlyCursorSeed(t *testing.T) {
	executor := &stubAllocationExecutor{
		queryRowResponses: []row{
			valueRow{values: []any{int64(21)}},
			valueRow{values: []any{int64(144)}},
		},
	}
	store := NewPaymentAddressAllocationStore(executor)
	input := newCloudflareReservePaymentAddressAllocationInput(" order-2 ")

	allocation, err := store.ReserveFresh(context.Background(), input)
	if err != nil {
		t.Fatalf("ReserveFresh returned error: %v", err)
	}
	if allocation.PaymentAddressID != 144 {
		t.Fatalf("unexpected payment address id: got %d", allocation.PaymentAddressID)
	}
	if allocation.DerivationIndex != 21 {
		t.Fatalf("unexpected derivation index: got %d", allocation.DerivationIndex)
	}
	if len(executor.execCalls) < 2 {
		t.Fatalf("expected at least 2 exec calls, got %d", len(executor.execCalls))
	}
	firstExec := executor.execCalls[0]
	if strings.Contains(firstExec.query, "address_source_ref IS NOT NULL") {
		t.Fatalf("did not expect legacy transitional seed branch, got %q", firstExec.query)
	}
	if !strings.Contains(firstExec.query, "ON CONFLICT (address_policy_id, address_source_ref) DO NOTHING") {
		t.Fatalf("expected xpub-backed conflict target, got %q", firstExec.query)
	}
	if len(firstExec.args) != 2 || firstExec.args[1] != "xpub-main" {
		t.Fatalf("unexpected cursor insert args: %+v", firstExec.args)
	}
	if len(executor.queryRowCalls) < 2 {
		t.Fatalf("expected at least 2 query row calls, got %d", len(executor.queryRowCalls))
	}
	if !strings.Contains(executor.queryRowCalls[0].query, "address_source_ref = $2") {
		t.Fatalf("expected cursor lookup by address_source_ref, got %q", executor.queryRowCalls[0].query)
	}
	if !strings.Contains(executor.queryRowCalls[1].query, "VALUES ($1, $2, $3, $4, $5, 'reserved')") {
		t.Fatalf("expected reserved insert values shape, got %q", executor.queryRowCalls[1].query)
	}
}
