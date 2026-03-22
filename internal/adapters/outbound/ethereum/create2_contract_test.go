package ethereum

import "testing"

func TestBuildFixedCollectorReceiverInitCodeHex(t *testing.T) {
	initCodeHex, err := BuildFixedCollectorReceiverInitCodeHex(
		"0x60006000556001600055",
		"0x2222222222222222222222222222222222222222",
	)
	if err != nil {
		t.Fatalf("BuildFixedCollectorReceiverInitCodeHex returned error: %v", err)
	}

	want := "0x600060005560016000550000000000000000000000002222222222222222222222222222222222222222"
	if initCodeHex != want {
		t.Fatalf("unexpected init code: got %q want %q", initCodeHex, want)
	}
}

func TestKeccak256Hex(t *testing.T) {
	got, err := Keccak256Hex("0x60006000556001600055")
	if err != nil {
		t.Fatalf("Keccak256Hex returned error: %v", err)
	}
	if got != "0x2afd5916fd398647ada97039811004dea871ae930e63202fd3beb21a751f188c" {
		t.Fatalf("unexpected keccak: got %q", got)
	}
}
