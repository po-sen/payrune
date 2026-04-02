package policies

import (
	"strings"
	"time"

	"payrune/internal/domain/valueobjects"
)

const (
	defaultPaymentReceiptRequiredConfirmations int32 = 1
	defaultPaymentReceiptExpiresAfter                = 7 * 24 * time.Hour
)

type PaymentAddressAllocationReservationAttempt string

const (
	PaymentAddressAllocationReservationAttemptReopenFailed PaymentAddressAllocationReservationAttempt = "reopen_failed"
	PaymentAddressAllocationReservationAttemptReserveFresh PaymentAddressAllocationReservationAttempt = "reserve_fresh"
)

type PaymentAddressAllocationReservation struct {
	IssuancePolicy      AddressIssuancePolicy
	ExpectedAmountMinor int64
	CustomerReference   string
}

type PaymentReceiptIssuanceTerms struct {
	RequiredConfirmations int32
	ExpiresAt             time.Time
}

type PaymentAddressAllocationIssuancePlan struct {
	Reservation         PaymentAddressAllocationReservation
	ReservationAttempts []PaymentAddressAllocationReservationAttempt
	ReceiptTerms        PaymentReceiptIssuanceTerms
}

type PaymentReceiptTermsScope struct {
	Chain   valueobjects.SupportedChain
	Network valueobjects.NetworkID
}

type PaymentAddressAllocationIssuancePolicy struct {
	requiredConfirmationsByScope map[PaymentReceiptTermsScope]int32
	receiptExpiresAfterByScope   map[PaymentReceiptTermsScope]time.Duration
}

func NewPaymentAddressAllocationIssuancePolicy(
	requiredConfirmationsByScope map[PaymentReceiptTermsScope]int32,
	receiptExpiresAfterByScope map[PaymentReceiptTermsScope]time.Duration,
) PaymentAddressAllocationIssuancePolicy {
	confirmationsByScope := make(map[PaymentReceiptTermsScope]int32)
	for scope, confirmations := range requiredConfirmationsByScope {
		if confirmations <= 0 {
			continue
		}
		normalizedScope, ok := normalizePaymentReceiptTermsScope(scope)
		if !ok {
			continue
		}
		confirmationsByScope[normalizedScope] = confirmations
	}

	expiresAfterByScope := make(map[PaymentReceiptTermsScope]time.Duration)
	for scope, expiresAfter := range receiptExpiresAfterByScope {
		if expiresAfter <= 0 {
			continue
		}
		normalizedScope, ok := normalizePaymentReceiptTermsScope(scope)
		if !ok {
			continue
		}
		expiresAfterByScope[normalizedScope] = expiresAfter
	}

	return PaymentAddressAllocationIssuancePolicy{
		requiredConfirmationsByScope: confirmationsByScope,
		receiptExpiresAfterByScope:   expiresAfterByScope,
	}
}

func (p PaymentAddressAllocationIssuancePolicy) Plan(
	issuancePolicy AddressIssuancePolicy,
	requestedChain valueobjects.SupportedChain,
	expectedAmountMinor int64,
	customerReference string,
	issuedAt time.Time,
) (PaymentAddressAllocationIssuancePlan, error) {
	if issuedAt.IsZero() {
		return PaymentAddressAllocationIssuancePlan{}, ErrPaymentAddressAllocationIssuedAtRequired
	}

	validatedPolicy, err := issuancePolicy.ValidateForAllocationIssuance(requestedChain, expectedAmountMinor)
	if err != nil {
		return PaymentAddressAllocationIssuancePlan{}, err
	}

	issuedAtUTC := issuedAt.UTC()
	return PaymentAddressAllocationIssuancePlan{
		Reservation: PaymentAddressAllocationReservation{
			IssuancePolicy:      validatedPolicy,
			ExpectedAmountMinor: expectedAmountMinor,
			CustomerReference:   strings.TrimSpace(customerReference),
		},
		ReservationAttempts: []PaymentAddressAllocationReservationAttempt{
			PaymentAddressAllocationReservationAttemptReopenFailed,
			PaymentAddressAllocationReservationAttemptReserveFresh,
		},
		ReceiptTerms: PaymentReceiptIssuanceTerms{
			RequiredConfirmations: p.requiredConfirmationsForScope(
				validatedPolicy.Chain,
				validatedPolicy.Network,
			),
			ExpiresAt: issuedAtUTC.Add(p.receiptExpiresAfterForScope(
				validatedPolicy.Chain,
				validatedPolicy.Network,
			)),
		},
	}, nil
}

func (p PaymentAddressAllocationIssuancePolicy) requiredConfirmationsForScope(
	chain valueobjects.SupportedChain,
	network valueobjects.NetworkID,
) int32 {
	scope, ok := normalizePaymentReceiptTermsScope(PaymentReceiptTermsScope{
		Chain:   chain,
		Network: network,
	})
	if ok {
		if configured, found := p.requiredConfirmationsByScope[scope]; found && configured > 0 {
			return configured
		}
	}
	return defaultPaymentReceiptRequiredConfirmations
}

func (p PaymentAddressAllocationIssuancePolicy) receiptExpiresAfterForScope(
	chain valueobjects.SupportedChain,
	network valueobjects.NetworkID,
) time.Duration {
	scope, ok := normalizePaymentReceiptTermsScope(PaymentReceiptTermsScope{
		Chain:   chain,
		Network: network,
	})
	if ok {
		if configured, found := p.receiptExpiresAfterByScope[scope]; found && configured > 0 {
			return configured
		}
	}
	return defaultPaymentReceiptExpiresAfter
}

func normalizePaymentReceiptTermsScope(scope PaymentReceiptTermsScope) (PaymentReceiptTermsScope, bool) {
	normalizedChain, ok := valueobjects.ParseSupportedChain(string(scope.Chain))
	if !ok {
		return PaymentReceiptTermsScope{}, false
	}
	normalizedNetwork, ok := valueobjects.ParseNetworkID(string(scope.Network))
	if !ok {
		return PaymentReceiptTermsScope{}, false
	}

	return PaymentReceiptTermsScope{
		Chain:   normalizedChain,
		Network: normalizedNetwork,
	}, true
}
