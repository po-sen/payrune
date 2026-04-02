package policies

import (
	"testing"

	"payrune/internal/domain/valueobjects"
)

func TestPollablePaymentReceiptStatuses(t *testing.T) {
	statuses := PollablePaymentReceiptStatuses()
	if len(statuses) != 4 {
		t.Fatalf("unexpected status count: got %d", len(statuses))
	}
	if statuses[0] != valueobjects.PaymentReceiptStatusWatching {
		t.Fatalf("unexpected first status: got %q", statuses[0])
	}
	if statuses[3] != valueobjects.PaymentReceiptStatusPaidUnconfirmedReverted {
		t.Fatalf("unexpected reverted status position: got %q", statuses[3])
	}
}
