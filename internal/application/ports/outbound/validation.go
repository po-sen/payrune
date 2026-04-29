package outbound

import (
	"strings"
	"time"
)

const (
	SupportedChainBitcoin  = "bitcoin"
	SupportedChainEthereum = "ethereum"

	NetworkIDMainnet  = "mainnet"
	NetworkIDTestnet4 = "testnet4"
	NetworkIDSepolia  = "sepolia"

	AddressSchemeLegacy       = "legacy"
	AddressSchemeSegwit       = "segwit"
	AddressSchemeNativeSegwit = "nativeSegwit"
	AddressSchemeTaproot      = "taproot"
	AddressSchemeCreate2      = "create2"

	IssuanceRefKindHDPathAbsolute = "hd_path_absolute"
	IssuanceRefKindCreate2Salt    = "create2_salt"

	PaymentAddressAllocationStatusReserved         = "reserved"
	PaymentAddressAllocationStatusIssued           = "issued"
	PaymentAddressAllocationStatusDerivationFailed = "derivation_failed"

	PaymentAddressAllocationFailureDerivationFailed = "derivation_failed"

	PaymentReceiptStatusWatching                = "watching"
	PaymentReceiptStatusPartiallyPaid           = "partially_paid"
	PaymentReceiptStatusPaidUnconfirmed         = "paid_unconfirmed"
	PaymentReceiptStatusPaidUnconfirmedReverted = "paid_unconfirmed_reverted"
	PaymentReceiptStatusPaidConfirmed           = "paid_confirmed"
	PaymentReceiptStatusFailedExpired           = "failed_expired"

	PaymentReceiptTrackingFailureReasonTrackingInvalid              = "tracking_invalid"
	PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable = "latest_block_height_unavailable"
	PaymentReceiptTrackingFailureReasonObservationFailed            = "observation_failed"
	PaymentReceiptTrackingFailureReasonTrackingUpdateFailed         = "tracking_update_failed"
	PaymentReceiptTrackingFailureReasonPaymentWindowExpired         = "payment_window_expired"
	PaymentReceiptTrackingFailureReasonProcessingFailed             = "processing_failed"

	PaymentReceiptNotificationDeliveryStatusPending = "pending"
	PaymentReceiptNotificationDeliveryStatusSent    = "sent"
	PaymentReceiptNotificationDeliveryStatusFailed  = "failed"

	PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed = "delivery_failed"
)

var addressSchemes = map[string]string{
	"legacy":       AddressSchemeLegacy,
	"segwit":       AddressSchemeSegwit,
	"nativesegwit": AddressSchemeNativeSegwit,
	"taproot":      AddressSchemeTaproot,
	"create2":      AddressSchemeCreate2,
}

var issuanceRefKinds = map[string]string{
	IssuanceRefKindHDPathAbsolute: IssuanceRefKindHDPathAbsolute,
	IssuanceRefKindCreate2Salt:    IssuanceRefKindCreate2Salt,
}

var paymentReceiptStatuses = map[string]string{
	PaymentReceiptStatusWatching:                PaymentReceiptStatusWatching,
	PaymentReceiptStatusPartiallyPaid:           PaymentReceiptStatusPartiallyPaid,
	PaymentReceiptStatusPaidUnconfirmed:         PaymentReceiptStatusPaidUnconfirmed,
	PaymentReceiptStatusPaidUnconfirmedReverted: PaymentReceiptStatusPaidUnconfirmedReverted,
	PaymentReceiptStatusPaidConfirmed:           PaymentReceiptStatusPaidConfirmed,
	PaymentReceiptStatusFailedExpired:           PaymentReceiptStatusFailedExpired,
}

var paymentReceiptTrackingFailureReasons = map[string]string{
	PaymentReceiptTrackingFailureReasonTrackingInvalid:              PaymentReceiptTrackingFailureReasonTrackingInvalid,
	PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable: PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable,
	PaymentReceiptTrackingFailureReasonObservationFailed:            PaymentReceiptTrackingFailureReasonObservationFailed,
	PaymentReceiptTrackingFailureReasonTrackingUpdateFailed:         PaymentReceiptTrackingFailureReasonTrackingUpdateFailed,
	PaymentReceiptTrackingFailureReasonPaymentWindowExpired:         PaymentReceiptTrackingFailureReasonPaymentWindowExpired,
	PaymentReceiptTrackingFailureReasonProcessingFailed:             PaymentReceiptTrackingFailureReasonProcessingFailed,
}

var notificationDeliveryStatuses = map[string]string{
	PaymentReceiptNotificationDeliveryStatusPending: PaymentReceiptNotificationDeliveryStatusPending,
	PaymentReceiptNotificationDeliveryStatusSent:    PaymentReceiptNotificationDeliveryStatusSent,
	PaymentReceiptNotificationDeliveryStatusFailed:  PaymentReceiptNotificationDeliveryStatusFailed,
}

var notificationDeliveryFailureReasons = map[string]string{
	PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed: PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed,
}

func NormalizeAddressPolicyID(raw string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	return normalized, isPortableID(normalized)
}

func NormalizeSupportedChain(raw string) (string, bool) {
	chain, ok := NormalizeChainID(raw)
	if !ok {
		return "", false
	}
	switch chain {
	case SupportedChainBitcoin, SupportedChainEthereum:
		return chain, true
	default:
		return "", false
	}
}

func NormalizeChainID(raw string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	return normalized, isPortableID(normalized)
}

func NormalizeNetworkID(raw string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	return normalized, isPortableID(normalized)
}

func NormalizeAddressScheme(raw string) (string, bool) {
	scheme, ok := addressSchemes[strings.ToLower(strings.TrimSpace(raw))]
	return scheme, ok
}

func NormalizeIssuanceRefKind(raw string) (string, bool) {
	kind, ok := issuanceRefKinds[strings.ToLower(strings.TrimSpace(raw))]
	return kind, ok
}

func NormalizePaymentReceiptStatus(raw string) (string, bool) {
	status, ok := paymentReceiptStatuses[strings.ToLower(strings.TrimSpace(raw))]
	return status, ok
}

func NormalizePaymentReceiptTrackingFailureReason(raw string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", false
	}
	reason, ok := paymentReceiptTrackingFailureReasons[normalized]
	return reason, ok
}

func NormalizePaymentAddressAllocationDerivationFailureReason(raw string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", false
	}
	if normalized == PaymentAddressAllocationFailureDerivationFailed {
		return normalized, true
	}
	return "", false
}

func NormalizePaymentReceiptNotificationDeliveryStatus(raw string) (string, bool) {
	status, ok := notificationDeliveryStatuses[strings.ToLower(strings.TrimSpace(raw))]
	return status, ok
}

func NormalizePaymentReceiptNotificationDeliveryFailureReason(raw string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", false
	}
	reason, ok := notificationDeliveryFailureReasons[normalized]
	return reason, ok
}

func PaymentReceiptTrackingFailureReasonMessage(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case PaymentReceiptTrackingFailureReasonTrackingInvalid:
		return "receipt tracking is invalid"
	case PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable:
		return "latest block height unavailable"
	case PaymentReceiptTrackingFailureReasonObservationFailed:
		return "receipt observation failed"
	case PaymentReceiptTrackingFailureReasonTrackingUpdateFailed:
		return "receipt tracking update failed"
	case PaymentReceiptTrackingFailureReasonPaymentWindowExpired:
		return "payment window expired"
	case PaymentReceiptTrackingFailureReasonProcessingFailed:
		return "receipt processing failed"
	default:
		return ""
	}
}

func PaymentReceiptNotificationDeliveryFailureReasonMessage(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed:
		return "receipt webhook delivery failed"
	default:
		return ""
	}
}

func MarkPaymentReceiptStatusNotificationSent(
	notificationID int64,
	deliveredAt time.Time,
) (PaymentReceiptStatusNotificationDeliveryResult, error) {
	if notificationID <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationOutboxFailed
	}
	if deliveredAt.IsZero() {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationDeliveredAtRequired
	}

	deliveredAtUTC := deliveredAt.UTC()
	return PaymentReceiptStatusNotificationDeliveryResult{
		NotificationID: notificationID,
		Status:         PaymentReceiptNotificationDeliveryStatusSent,
		DeliveredAt:    &deliveredAtUTC,
	}, nil
}

func ResolvePaymentReceiptStatusNotificationDeliveryFailure(
	notificationID int64,
	currentAttempts int32,
	maxAttempts int32,
	now time.Time,
	retryDelay time.Duration,
	failureReason string,
) (PaymentReceiptStatusNotificationDeliveryResult, error) {
	if notificationID <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationOutboxFailed
	}
	if currentAttempts < 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationOutboxFailed
	}
	if maxAttempts <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationOutboxFailed
	}
	if now.IsZero() {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationOutboxFailed
	}
	normalizedFailureReason, ok := NormalizePaymentReceiptNotificationDeliveryFailureReason(failureReason)
	if !ok {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationOutboxFailed
	}

	attempts := currentAttempts + 1
	if attempts >= maxAttempts {
		return PaymentReceiptStatusNotificationDeliveryResult{
			NotificationID:    notificationID,
			Status:            PaymentReceiptNotificationDeliveryStatusFailed,
			Attempts:          attempts,
			LastFailureReason: normalizedFailureReason,
		}, nil
	}

	if retryDelay <= 0 {
		return PaymentReceiptStatusNotificationDeliveryResult{}, ErrPaymentReceiptStatusNotificationOutboxFailed
	}

	nextAttemptAt := now.Add(retryDelay).UTC()
	return PaymentReceiptStatusNotificationDeliveryResult{
		NotificationID:    notificationID,
		Status:            PaymentReceiptNotificationDeliveryStatusPending,
		Attempts:          attempts,
		LastFailureReason: normalizedFailureReason,
		NextAttemptAt:     &nextAttemptAt,
	}, nil
}

func isPortableID(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for i := 0; i < len(value); i++ {
		char := value[i]
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '_' || char == '-' {
			continue
		}
		return false
	}
	return true
}
