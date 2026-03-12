package policies

import (
	"errors"
	"strings"
	"time"

	"payrune/internal/domain/entities"
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
	IssuancePolicy      entities.AddressIssuancePolicy
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

type PaymentAddressAllocationIssuancePolicy struct {
	requiredConfirmationsByNetwork map[valueobjects.NetworkID]int32
	receiptExpiresAfterByNetwork   map[valueobjects.NetworkID]time.Duration
}

func NewPaymentAddressAllocationIssuancePolicy(
	requiredConfirmationsByNetwork map[valueobjects.NetworkID]int32,
	receiptExpiresAfterByNetwork map[valueobjects.NetworkID]time.Duration,
) PaymentAddressAllocationIssuancePolicy {
	confirmationsByNetwork := make(map[valueobjects.NetworkID]int32)
	for network, confirmations := range requiredConfirmationsByNetwork {
		if confirmations <= 0 {
			continue
		}
		confirmationsByNetwork[network] = confirmations
	}

	expiresAfterByNetwork := make(map[valueobjects.NetworkID]time.Duration)
	for network, expiresAfter := range receiptExpiresAfterByNetwork {
		if expiresAfter <= 0 {
			continue
		}
		expiresAfterByNetwork[network] = expiresAfter
	}

	return PaymentAddressAllocationIssuancePolicy{
		requiredConfirmationsByNetwork: confirmationsByNetwork,
		receiptExpiresAfterByNetwork:   expiresAfterByNetwork,
	}
}

func (p PaymentAddressAllocationIssuancePolicy) Plan(
	issuancePolicy entities.AddressIssuancePolicy,
	requestedChain valueobjects.SupportedChain,
	expectedAmountMinor int64,
	customerReference string,
	issuedAt time.Time,
) (PaymentAddressAllocationIssuancePlan, error) {
	if issuedAt.IsZero() {
		return PaymentAddressAllocationIssuancePlan{}, errors.New("issued at is required")
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
			RequiredConfirmations: p.requiredConfirmationsForNetwork(validatedPolicy.AddressPolicy.Network),
			ExpiresAt:             issuedAtUTC.Add(p.receiptExpiresAfterForNetwork(validatedPolicy.AddressPolicy.Network)),
		},
	}, nil
}

func (p PaymentAddressAllocationIssuancePolicy) requiredConfirmationsForNetwork(
	network valueobjects.NetworkID,
) int32 {
	if configured, ok := p.requiredConfirmationsByNetwork[network]; ok && configured > 0 {
		return configured
	}
	return defaultPaymentReceiptRequiredConfirmations
}

func (p PaymentAddressAllocationIssuancePolicy) receiptExpiresAfterForNetwork(
	network valueobjects.NetworkID,
) time.Duration {
	if configured, ok := p.receiptExpiresAfterByNetwork[network]; ok && configured > 0 {
		return configured
	}
	return defaultPaymentReceiptExpiresAfter
}
