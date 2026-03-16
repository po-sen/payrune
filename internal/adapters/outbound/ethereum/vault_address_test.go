package ethereum

import (
	"bytes"
	"testing"
)

func TestGenerateSaltHex(t *testing.T) {
	saltHex, err := GenerateSaltHex(bytes.NewReader(bytes.Repeat([]byte{0x11}, 32)))
	if err != nil {
		t.Fatalf("GenerateSaltHex returned error: %v", err)
	}
	want := "0x1111111111111111111111111111111111111111111111111111111111111111"
	if saltHex != want {
		t.Fatalf("unexpected salt hex: got %q want %q", saltHex, want)
	}
}

func TestPredictVaultAddress(t *testing.T) {
	address, err := PredictVaultAddress(
		"0x3333333333333333333333333333333333333333",
		"0x1111111111111111111111111111111111111111111111111111111111111111",
		"0x4023b4c6529dc6ac7f0b3db124a2b8c86febf4ab1e6e491612ca4a068fae6e21",
	)
	if err != nil {
		t.Fatalf("PredictVaultAddress returned error: %v", err)
	}
	want := "0x5816b1fbecac478596e7436d2cee2cf37071f47b"
	if address != want {
		t.Fatalf("unexpected predicted address: got %q want %q", address, want)
	}
}
