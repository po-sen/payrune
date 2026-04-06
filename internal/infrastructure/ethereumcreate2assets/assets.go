package ethereumcreate2assets

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

//go:embed artifacts/*.json metadata/*.json
var embeddedAssets embed.FS

const (
	ReceiverArtifactName = "FixedCollectorReceiver.json"
	FactoryArtifactName  = "Create2ReceiverFactory.json"
)

type DeploymentMetadata struct {
	Network          string           `json:"network"`
	FactoryAddress   string           `json:"factoryAddress"`
	ReceiverArtifact string           `json:"receiverArtifact"`
	Mode             string           `json:"mode,omitempty"`
	Receiver         ReceiverArtifact `json:"-"`
}

type ReceiverArtifact struct {
	SourceName      string          `json:"sourceName"`
	ContractName    string          `json:"contractName"`
	CompilerVersion string          `json:"compilerVersion"`
	ABI             json.RawMessage `json:"abi"`
	CreationCodeHex string          `json:"creationCodeHex"`
	RuntimeCodeHex  string          `json:"runtimeCodeHex"`
}

var deploymentMetadataByNetwork = mustLoadDeploymentMetadataByNetwork()
var receiverArtifactsByName = mustLoadReceiverArtifactsByName()

func LookupDeploymentMetadata(network string) (DeploymentMetadata, bool) {
	metadata, ok := deploymentMetadataByNetwork[normalizeNetworkKey(network)]
	return metadata, ok
}

func BuildAddressSpaceRef(network string, collectorAddress string) string {
	metadata, ok := LookupDeploymentMetadata(network)
	if !ok {
		return ""
	}
	return BuildAddressSpaceRefFromMetadata(metadata, collectorAddress)
}

func BuildTokenCapableAddressSpaceRef(network string, collectorAddress string) string {
	return BuildAddressSpaceRefWithReceiverArtifact(network, collectorAddress, ReceiverArtifactName)
}

func BuildAddressSpaceRefWithReceiverArtifact(
	network string,
	collectorAddress string,
	receiverArtifactName string,
) string {
	metadata, ok := LookupDeploymentMetadata(network)
	if !ok {
		return ""
	}

	artifact, ok := LookupReceiverArtifact(receiverArtifactName)
	if !ok {
		return ""
	}

	metadata.ReceiverArtifact = receiverArtifactName
	metadata.Receiver = artifact
	return BuildAddressSpaceRefFromMetadata(metadata, collectorAddress)
}

func BuildAddressSpaceRefFromMetadata(
	metadata DeploymentMetadata,
	collectorAddress string,
) string {
	collectorAddress = strings.TrimSpace(collectorAddress)
	if collectorAddress == "" {
		return ""
	}

	initCodeHash, ok := metadata.Receiver.InitCodeHashHex(collectorAddress)
	if !ok {
		return ""
	}

	sourceRef, err := buildCreate2AddressSpaceRef(
		strings.TrimSpace(metadata.FactoryAddress),
		collectorAddress,
		initCodeHash,
	)
	if err != nil {
		return ""
	}
	return sourceRef
}

func mustLoadDeploymentMetadataByNetwork() map[string]DeploymentMetadata {
	entries, err := embeddedAssets.ReadDir("metadata")
	if err != nil {
		panic(fmt.Errorf("read embedded ethereum create2 metadata: %w", err))
	}

	loaded := make(map[string]DeploymentMetadata, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		metadata, err := loadDeploymentMetadata(entry.Name())
		if err != nil {
			panic(err)
		}

		networkKey := normalizeNetworkKey(metadata.Network)
		if networkKey == "" {
			panic(fmt.Errorf("embedded ethereum create2 metadata network is invalid: %s", metadata.Network))
		}
		metadata.Network = networkKey
		if _, exists := loaded[networkKey]; exists {
			panic(fmt.Errorf("duplicate embedded ethereum create2 metadata for network: %s", networkKey))
		}
		loaded[networkKey] = metadata
	}

	return loaded
}

func mustLoadReceiverArtifactsByName() map[string]ReceiverArtifact {
	entries, err := embeddedAssets.ReadDir("artifacts")
	if err != nil {
		panic(fmt.Errorf("read embedded ethereum create2 artifacts: %w", err))
	}

	loaded := make(map[string]ReceiverArtifact, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		artifact, err := loadReceiverArtifact(entry.Name())
		if err != nil {
			panic(err)
		}
		if _, exists := loaded[entry.Name()]; exists {
			panic(fmt.Errorf("duplicate embedded ethereum create2 receiver artifact: %s", entry.Name()))
		}
		loaded[entry.Name()] = artifact
	}

	return loaded
}

func LookupReceiverArtifact(fileName string) (ReceiverArtifact, bool) {
	artifact, ok := receiverArtifactsByName[strings.TrimSpace(fileName)]
	return artifact, ok
}

func loadDeploymentMetadata(fileName string) (DeploymentMetadata, error) {
	raw, err := embeddedAssets.ReadFile(path.Join("metadata", fileName))
	if err != nil {
		return DeploymentMetadata{}, fmt.Errorf("read ethereum create2 metadata %s: %w", fileName, err)
	}

	var metadata DeploymentMetadata
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return DeploymentMetadata{}, fmt.Errorf("decode ethereum create2 metadata %s: %w", fileName, err)
	}
	metadata.FactoryAddress = strings.TrimSpace(metadata.FactoryAddress)
	metadata.ReceiverArtifact = strings.TrimSpace(metadata.ReceiverArtifact)
	metadata.Network = normalizeNetworkKey(metadata.Network)
	if metadata.FactoryAddress == "" || metadata.ReceiverArtifact == "" || metadata.Network == "" {
		return DeploymentMetadata{}, fmt.Errorf("ethereum create2 metadata %s is incomplete", fileName)
	}

	artifact, err := loadReceiverArtifact(metadata.ReceiverArtifact)
	if err != nil {
		return DeploymentMetadata{}, err
	}
	metadata.Receiver = artifact
	return metadata, nil
}

func loadReceiverArtifact(fileName string) (ReceiverArtifact, error) {
	raw, err := embeddedAssets.ReadFile(path.Join("artifacts", fileName))
	if err != nil {
		return ReceiverArtifact{}, fmt.Errorf("read ethereum create2 receiver artifact %s: %w", fileName, err)
	}

	var artifact ReceiverArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return ReceiverArtifact{}, fmt.Errorf("decode ethereum create2 receiver artifact %s: %w", fileName, err)
	}
	if strings.TrimSpace(artifact.CreationCodeHex) == "" {
		return ReceiverArtifact{}, fmt.Errorf("ethereum create2 receiver artifact %s is incomplete", fileName)
	}
	return artifact, nil
}

func (a ReceiverArtifact) InitCodeHex(collectorAddress string) (string, bool) {
	initCodeHex, err := buildFixedCollectorReceiverInitCodeHex(
		a.CreationCodeHex,
		collectorAddress,
	)
	if err != nil {
		return "", false
	}
	return initCodeHex, true
}

func (a ReceiverArtifact) InitCodeHashHex(collectorAddress string) (string, bool) {
	initCodeHex, ok := a.InitCodeHex(collectorAddress)
	if !ok {
		return "", false
	}

	initCodeHashHex, err := keccak256Hex(initCodeHex)
	if err != nil {
		return "", false
	}
	return initCodeHashHex, true
}

func normalizeNetworkKey(network string) string {
	return strings.ToLower(strings.TrimSpace(network))
}
