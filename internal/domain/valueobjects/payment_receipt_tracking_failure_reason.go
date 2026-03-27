package valueobjects

import "strings"

type PaymentReceiptTrackingFailureReason string

const (
	PaymentReceiptTrackingFailureReasonTrackingInvalid              PaymentReceiptTrackingFailureReason = "tracking_invalid"
	PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable PaymentReceiptTrackingFailureReason = "latest_block_height_unavailable"
	PaymentReceiptTrackingFailureReasonObservationFailed            PaymentReceiptTrackingFailureReason = "observation_failed"
	PaymentReceiptTrackingFailureReasonTrackingUpdateFailed         PaymentReceiptTrackingFailureReason = "tracking_update_failed"
	PaymentReceiptTrackingFailureReasonPaymentWindowExpired         PaymentReceiptTrackingFailureReason = "payment_window_expired"
	PaymentReceiptTrackingFailureReasonProcessingFailed             PaymentReceiptTrackingFailureReason = "processing_failed"
)

var paymentReceiptTrackingFailureReasons = map[string]PaymentReceiptTrackingFailureReason{
	"tracking_invalid":                PaymentReceiptTrackingFailureReasonTrackingInvalid,
	"latest_block_height_unavailable": PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable,
	"observation_failed":              PaymentReceiptTrackingFailureReasonObservationFailed,
	"tracking_update_failed":          PaymentReceiptTrackingFailureReasonTrackingUpdateFailed,
	"payment_window_expired":          PaymentReceiptTrackingFailureReasonPaymentWindowExpired,
	"processing_failed":               PaymentReceiptTrackingFailureReasonProcessingFailed,
}

var paymentReceiptTrackingFailureReasonLegacyAliases = map[string]PaymentReceiptTrackingFailureReason{
	"receipt tracking is invalid":     PaymentReceiptTrackingFailureReasonTrackingInvalid,
	"issued at is required":           PaymentReceiptTrackingFailureReasonTrackingInvalid,
	"latest block height unavailable": PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable,
	"tip height timeout":              PaymentReceiptTrackingFailureReasonLatestBlockHeightUnavailable,
	"receipt observation failed":      PaymentReceiptTrackingFailureReasonObservationFailed,
	"rpc timeout":                     PaymentReceiptTrackingFailureReasonObservationFailed,
	"receipt tracking update failed":  PaymentReceiptTrackingFailureReasonTrackingUpdateFailed,
	"payment window expired":          PaymentReceiptTrackingFailureReasonPaymentWindowExpired,
	"receipt processing failed":       PaymentReceiptTrackingFailureReasonProcessingFailed,
}

func ParsePaymentReceiptTrackingFailureReason(raw string) (PaymentReceiptTrackingFailureReason, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", false
	}
	if reason, ok := paymentReceiptTrackingFailureReasons[normalized]; ok {
		return reason, true
	}
	if reason, ok := paymentReceiptTrackingFailureReasonLegacyAliases[normalized]; ok {
		return reason, true
	}
	return PaymentReceiptTrackingFailureReasonProcessingFailed, true
}

func (r PaymentReceiptTrackingFailureReason) IsZero() bool {
	return strings.TrimSpace(string(r)) == ""
}

func (r PaymentReceiptTrackingFailureReason) Message() string {
	switch r {
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
