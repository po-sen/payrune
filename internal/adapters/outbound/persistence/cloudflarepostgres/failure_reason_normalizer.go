package cloudflarepostgres

import (
	"strings"

	applicationoutbox "payrune/internal/application/outbox"
	"payrune/internal/domain/valueobjects"
)

func normalizePaymentReceiptTrackingFailureReason(raw string) valueobjects.PaymentReceiptTrackingFailureReason {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}
	if reason, ok := valueobjects.ParsePaymentReceiptTrackingFailureReason(normalized); ok {
		return reason
	}

	switch normalized {
	case "receipt tracking is invalid", "issued at is required":
		return valueobjects.PaymentReceiptTrackingFailureReasonTrackingInvalid
	case "latest block height unavailable", "tip height timeout":
		return valueobjects.PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable
	case "receipt observation failed", "rpc timeout":
		return valueobjects.PaymentReceiptTrackingFailureReasonObservationFailed
	case "receipt tracking update failed":
		return valueobjects.PaymentReceiptTrackingFailureReasonTrackingUpdateFailed
	case "payment window expired":
		return valueobjects.PaymentReceiptTrackingFailureReasonPaymentWindowExpired
	case "receipt processing failed":
		return valueobjects.PaymentReceiptTrackingFailureReasonProcessingFailed
	default:
		return valueobjects.PaymentReceiptTrackingFailureReasonProcessingFailed
	}
}

func normalizePaymentReceiptNotificationDeliveryFailureReason(
	raw string,
) applicationoutbox.PaymentReceiptNotificationDeliveryFailureReason {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}
	if reason, ok := applicationoutbox.ParsePaymentReceiptNotificationDeliveryFailureReason(normalized); ok {
		return reason
	}

	switch normalized {
	case "receipt webhook delivery failed", "timeout", "webhook returned status 500":
		return applicationoutbox.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed
	default:
		return applicationoutbox.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed
	}
}

func normalizePaymentAddressAllocationDerivationFailureReason(
	raw string,
) valueobjects.PaymentAddressAllocationDerivationFailureReason {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}
	if reason, ok := valueobjects.ParsePaymentAddressAllocationDerivationFailureReason(normalized); ok {
		return reason
	}

	switch normalized {
	case "derive failed", "derivation failed", "address derivation failed", "payment address derivation failed":
		return valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed
	default:
		return valueobjects.PaymentAddressAllocationDerivationFailureReasonDerivationFailed
	}
}
