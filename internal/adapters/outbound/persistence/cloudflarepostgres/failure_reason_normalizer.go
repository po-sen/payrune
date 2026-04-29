package cloudflarepostgres

import (
	"strings"

	outport "payrune/internal/application/ports/outbound"
)

func normalizePaymentReceiptTrackingFailureReason(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}
	if reason, ok := outport.NormalizePaymentReceiptTrackingFailureReason(normalized); ok {
		return reason
	}

	switch normalized {
	case "receipt tracking is invalid", "issued at is required":
		return outport.PaymentReceiptTrackingFailureReasonTrackingInvalid
	case "latest block height unavailable", "tip height timeout":
		return outport.PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable
	case "receipt observation failed", "rpc timeout":
		return outport.PaymentReceiptTrackingFailureReasonObservationFailed
	case "receipt tracking update failed":
		return outport.PaymentReceiptTrackingFailureReasonTrackingUpdateFailed
	case "payment window expired":
		return outport.PaymentReceiptTrackingFailureReasonPaymentWindowExpired
	case "receipt processing failed":
		return outport.PaymentReceiptTrackingFailureReasonProcessingFailed
	default:
		return outport.PaymentReceiptTrackingFailureReasonProcessingFailed
	}
}

func normalizePaymentReceiptNotificationDeliveryFailureReason(
	raw string,
) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}
	if reason, ok := outport.NormalizePaymentReceiptNotificationDeliveryFailureReason(normalized); ok {
		return reason
	}

	switch normalized {
	case "receipt webhook delivery failed", "timeout", "webhook returned status 500":
		return outport.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed
	default:
		return outport.PaymentReceiptNotificationDeliveryFailureReasonDeliveryFailed
	}
}

func normalizePaymentAddressAllocationDerivationFailureReason(
	raw string,
) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}
	if reason, ok := outport.NormalizePaymentAddressAllocationDerivationFailureReason(normalized); ok {
		return reason
	}

	switch normalized {
	case "derive failed", "derivation failed", "address derivation failed", "payment address derivation failed":
		return outport.PaymentAddressAllocationFailureDerivationFailed
	default:
		return outport.PaymentAddressAllocationFailureDerivationFailed
	}
}
