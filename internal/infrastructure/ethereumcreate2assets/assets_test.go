package ethereumcreate2assets

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"golang.org/x/crypto/sha3"
)

func TestBuildAddressSpaceRefFromMetadataRequiresCompleteInputs(t *testing.T) {
	collectorAddress := "0x2222222222222222222222222222222222222222"

	if got := BuildAddressSpaceRefFromMetadata(
		DeploymentMetadata{},
		collectorAddress,
	); got != "" {
		t.Fatalf("expected empty source ref for missing metadata, got %q", got)
	}

	if got := BuildAddressSpaceRefFromMetadata(
		DeploymentMetadata{
			FactoryAddress: "0x1111111111111111111111111111111111111111",
		},
		collectorAddress,
	); got != "" {
		t.Fatalf("expected empty source ref for missing receiver artifact, got %q", got)
	}

	if got := BuildAddressSpaceRefFromMetadata(
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
	mainnet, ok := LookupDeploymentMetadata("mainnet")
	if !ok {
		t.Fatal("expected embedded mainnet metadata")
	}
	if mainnet.FactoryAddress == "" {
		t.Fatal("expected embedded mainnet factory address")
	}
	if strings.TrimSpace(mainnet.Receiver.CreationCodeHex) == "" {
		t.Fatal("expected embedded receiver artifact creation code")
	}

	sepolia, ok := LookupDeploymentMetadata("sepolia")
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

func TestEmbeddedFactoryArtifactLoads(t *testing.T) {
	artifact, err := loadReceiverArtifact("Create2ReceiverFactoryV1.json")
	if err != nil {
		t.Fatalf("loadReceiverArtifact returned error: %v", err)
	}
	if artifact.ContractName != "Create2ReceiverFactory" {
		t.Fatalf("unexpected contract name: got %q", artifact.ContractName)
	}
	if strings.TrimSpace(artifact.CreationCodeHex) == "" {
		t.Fatal("expected factory creation code")
	}
	if len(artifact.ABI) == 0 {
		t.Fatal("expected factory ABI")
	}

	var abiItems []map[string]any
	if err := json.Unmarshal(artifact.ABI, &abiItems); err != nil {
		t.Fatalf("json.Unmarshal ABI returned error: %v", err)
	}

	foundSweep := false
	foundDeployAndCall := false
	for _, item := range abiItems {
		typ, _ := item["type"].(string)
		name, _ := item["name"].(string)
		if typ != "function" {
			continue
		}
		if name == "sweep" {
			foundSweep = true
		}
		if name == "deployAndCall" {
			foundDeployAndCall = true
		}
	}
	if !foundSweep {
		t.Fatal("expected factory ABI to expose sweep")
	}
	if foundDeployAndCall {
		t.Fatal("did not expect factory ABI to expose deployAndCall")
	}
}
