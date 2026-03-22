package ethereumcreate2assets

import (
	"encoding/hex"
	"strings"
	"testing"

	"golang.org/x/crypto/sha3"

	"payrune/internal/domain/valueobjects"
)

func TestBuildAddressSourceRefFromMetadataRequiresCompleteInputs(t *testing.T) {
	collectorAddress := "0x2222222222222222222222222222222222222222"

	if got := BuildAddressSourceRefFromMetadata(
		DeploymentMetadata{},
		collectorAddress,
	); got != "" {
		t.Fatalf("expected empty source ref for missing metadata, got %q", got)
	}

	if got := BuildAddressSourceRefFromMetadata(
		DeploymentMetadata{
			FactoryAddress: "0x1111111111111111111111111111111111111111",
		},
		collectorAddress,
	); got != "" {
		t.Fatalf("expected empty source ref for missing receiver artifact, got %q", got)
	}

	if got := BuildAddressSourceRefFromMetadata(
		DeploymentMetadata{
			FactoryAddress: "0x1111111111111111111111111111111111111111",
		},
		"",
	); got != "" {
		t.Fatalf("expected empty source ref for missing collector, got %q", got)
	}
}

func TestReceiverArtifactInitCodeHashHex(t *testing.T) {
	collectorAddress := "0x2222222222222222222222222222222222222222"
	artifact := ReceiverArtifact{
		CreationCodeHex: "0x60006000556001600055",
	}

	got, ok := artifact.InitCodeHashHex(collectorAddress)
	if !ok {
		t.Fatal("expected init code hash available")
	}

	initCodeHex, ok := artifact.InitCodeHex(collectorAddress)
	if !ok {
		t.Fatal("expected init code available")
	}

	initCode, err := hex.DecodeString(initCodeHex[2:])
	if err != nil {
		t.Fatalf("DecodeString returned error: %v", err)
	}
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write(initCode)
	expected := "0x" + hex.EncodeToString(hasher.Sum(nil))

	if got != expected {
		t.Fatalf("unexpected init code hash: got %q want %q", got, expected)
	}
}

func TestEmbeddedMetadataLoadsMainnetAndSepolia(t *testing.T) {
	mainnet, ok := LookupDeploymentMetadata(valueobjects.NetworkID("mainnet"))
	if !ok {
		t.Fatal("expected embedded mainnet metadata")
	}
	if mainnet.FactoryAddress == "" {
		t.Fatal("expected embedded mainnet factory address")
	}
	if strings.TrimSpace(mainnet.Receiver.CreationCodeHex) == "" {
		t.Fatal("expected embedded receiver artifact creation code")
	}

	sepolia, ok := LookupDeploymentMetadata(valueobjects.NetworkID("sepolia"))
	if !ok {
		t.Fatal("expected embedded sepolia metadata")
	}
	if sepolia.FactoryAddress == "" {
		t.Fatal("expected embedded sepolia factory address")
	}
	if strings.TrimSpace(sepolia.Receiver.CreationCodeHex) == "" {
		t.Fatal("expected embedded receiver artifact creation code")
	}
}
