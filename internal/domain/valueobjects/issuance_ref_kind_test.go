package valueobjects

import "testing"

func TestParseIssuanceRefKind(t *testing.T) {
	t.Run("known kind", func(t *testing.T) {
		kind, ok := ParseIssuanceRefKind(" create2_salt ")
		if !ok {
			t.Fatal("expected kind to parse")
		}
		if kind != IssuanceRefKindCreate2Salt {
			t.Fatalf("unexpected kind: got %q", kind)
		}
	})

	t.Run("unknown kind", func(t *testing.T) {
		if _, ok := ParseIssuanceRefKind("unknown"); ok {
			t.Fatal("expected parse failure")
		}
	})
}

func TestIssuanceRefKindIsZero(t *testing.T) {
	if !(IssuanceRefKind(" ")).IsZero() {
		t.Fatal("expected blank kind to be zero")
	}
	if IssuanceRefKindHDPathAbsolute.IsZero() {
		t.Fatal("expected non-blank kind to be non-zero")
	}
}
