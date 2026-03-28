package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	"payrune/internal/domain/valueobjects"
)

func TestChainAddressControllerListSuccess(t *testing.T) {
	listUC := &fakeListAddressPoliciesUseCase{
		response: dto.ListAddressPoliciesResponse{
			Chain: "bitcoin",
			AddressPolicies: []dto.AddressPolicy{{
				AddressPolicyID: "bitcoin-mainnet-legacy",
				Chain:           "bitcoin",
				Network:         "mainnet",
				Scheme:          "legacy",
				MinorUnit:       "satoshi",
				Decimals:        8,
				Enabled:         true,
			}},
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/chains/{chain}/address-policies", NewListAddressPoliciesController(listUC))

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if listUC.lastChain != valueobjects.SupportedChainBitcoin {
		t.Fatalf("unexpected chain passed to use case: got %q", listUC.lastChain)
	}

	var body dto.ListAddressPoliciesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.AddressPolicies) != 1 {
		t.Fatalf("unexpected policy count: got %d", len(body.AddressPolicies))
	}
	if body.AddressPolicies[0].MinorUnit != "satoshi" {
		t.Fatalf("unexpected minor unit: got %q", body.AddressPolicies[0].MinorUnit)
	}
	if body.AddressPolicies[0].Decimals != 8 {
		t.Fatalf("unexpected decimals: got %d", body.AddressPolicies[0].Decimals)
	}
}

func TestChainAddressControllerListEthereumSuccess(t *testing.T) {
	listUC := &fakeListAddressPoliciesUseCase{
		response: dto.ListAddressPoliciesResponse{
			Chain: "ethereum",
			AddressPolicies: []dto.AddressPolicy{{
				AddressPolicyID: "ethereum-mainnet-create2",
				Chain:           "ethereum",
				Network:         "mainnet",
				Scheme:          "create2",
				MinorUnit:       "wei",
				Decimals:        18,
				Enabled:         true,
			}},
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/chains/{chain}/address-policies", NewListAddressPoliciesController(listUC))

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/ethereum/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if listUC.lastChain != valueobjects.SupportedChainEthereum {
		t.Fatalf("unexpected chain passed to use case: got %q", listUC.lastChain)
	}
}

func TestChainAddressControllerListRejectMethod(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/address-policies",
		NewListAddressPoliciesController(&fakeListAddressPoliciesUseCase{}),
	)

	req := httptest.NewRequest(http.MethodPost, "/v1/chains/bitcoin/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
	if allow := rr.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("unexpected Allow header: got %q", allow)
	}
}

func TestChainAddressControllerListInternalError(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(
		"/v1/chains/{chain}/address-policies",
		NewListAddressPoliciesController(&fakeListAddressPoliciesUseCase{err: inport.ErrDependencyFailure}),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/chains/bitcoin/address-policies", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}
}
