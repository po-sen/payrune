package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"payrune/internal/adapters/outbound/ethereum"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

const (
	defaultPredictNetwork = "mainnet"
	defaultSolcImageRef   = "ghcr.io/argotorg/solc@sha256:6263a14716bf74f01cc80e86e0fcd28a5bae4d4aca46cc8aa6f4c2d6608ab143"
	defaultSolcPlatform   = "linux/amd64"
)

type toolPaths struct {
	repoRoot         string
	assetsDir        string
	contractsDir     string
	artifactsDir     string
	factoryArtifact  string
	receiverArtifact string
}

type receiverArtifact struct {
	SourceName      string          `json:"sourceName"`
	ContractName    string          `json:"contractName"`
	CompilerVersion string          `json:"compilerVersion"`
	ABI             json.RawMessage `json:"abi"`
	CreationCodeHex string          `json:"creationCodeHex"`
	RuntimeCodeHex  string          `json:"runtimeCodeHex"`
}

type predictionOutput struct {
	SourceRef         string `json:"sourceRef"`
	InitCodeHex       string `json:"initCodeHex"`
	InitCodeHashHex   string `json:"initCodeHashHex"`
	PredictedAddress  string `json:"predictedAddress"`
	Salt              string `json:"salt"`
	AddressReference  string `json:"addressReference"`
	RelativeReference string `json:"relativeAddressReference"`
}

func resolveToolPaths() (toolPaths, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return toolPaths{}, err
	}

	repoRoot, err := findRepoRoot(cwd)
	if err != nil {
		return toolPaths{}, err
	}

	assetsDir := filepath.Join(repoRoot, "internal", "infrastructure", "ethereumcreate2assets")
	artifactsDir := filepath.Join(assetsDir, "artifacts")
	return toolPaths{
		repoRoot:         repoRoot,
		assetsDir:        assetsDir,
		contractsDir:     filepath.Join(assetsDir, "contracts"),
		artifactsDir:     artifactsDir,
		factoryArtifact:  filepath.Join(artifactsDir, "Create2ReceiverFactoryV1.json"),
		receiverArtifact: filepath.Join(artifactsDir, "FixedCollectorReceiverV1.json"),
	}, nil
}

func findRepoRoot(start string) (string, error) {
	current := filepath.Clean(start)
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("could not find repo root containing go.mod")
		}
		current = parent
	}
}

func loadReceiverArtifact(path string) (receiverArtifact, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return receiverArtifact{}, err
	}

	var artifact receiverArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return receiverArtifact{}, err
	}
	if strings.TrimSpace(artifact.CreationCodeHex) == "" {
		return receiverArtifact{}, errors.New("receiver artifact creationCodeHex is required")
	}
	if len(artifact.ABI) == 0 {
		return receiverArtifact{}, errors.New("receiver artifact abi is required")
	}
	return artifact, nil
}

func predictFromArtifact(
	ctx context.Context,
	network string,
	factoryAddress string,
	collectorAddress string,
	receiverArtifactPath string,
	addressPrefix string,
	salt string,
) (predictionOutput, error) {
	artifact, err := loadReceiverArtifact(receiverArtifactPath)
	if err != nil {
		return predictionOutput{}, err
	}

	initCodeHex, err := ethereum.BuildFixedCollectorReceiverInitCodeHex(
		artifact.CreationCodeHex,
		strings.TrimSpace(collectorAddress),
	)
	if err != nil {
		return predictionOutput{}, err
	}

	initCodeHashHex, err := ethereum.Keccak256Hex(initCodeHex)
	if err != nil {
		return predictionOutput{}, err
	}

	sourceRef, err := ethereum.BuildCreate2AddressSourceRef(
		strings.TrimSpace(factoryAddress),
		strings.TrimSpace(collectorAddress),
		initCodeHashHex,
	)
	if err != nil {
		return predictionOutput{}, err
	}

	deriver := ethereum.NewChainAddressDeriver()
	output, err := deriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:                    valueobjects.SupportedChainEthereum,
		Network:                  valueobjects.NetworkID(strings.TrimSpace(network)),
		Scheme:                   "create2",
		AddressSourceRef:         sourceRef,
		AddressReferencePrefix:   strings.TrimSpace(addressPrefix),
		RelativeAddressReference: strings.TrimSpace(salt),
	})
	if err != nil {
		return predictionOutput{}, err
	}

	return predictionOutput{
		SourceRef:         sourceRef,
		InitCodeHex:       initCodeHex,
		InitCodeHashHex:   initCodeHashHex,
		PredictedAddress:  output.Address,
		Salt:              output.RelativeAddressReference,
		AddressReference:  output.AddressReference,
		RelativeReference: output.RelativeAddressReference,
	}, nil
}

func newFlagSet(name string) *flag.FlagSet {
	flagSet := flag.NewFlagSet(name, flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	return flagSet
}

func writePrettyJSON(target *os.File, value any) error {
	encoder := json.NewEncoder(target)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func usageError(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func defaultAddressReferencePrefix(network string) string {
	trimmed := strings.TrimSpace(network)
	if trimmed == "" {
		return ""
	}
	return "ethereum-" + trimmed + "-create2"
}

func normalizeOrGenerateSalt(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		return raw, nil
	}

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(salt), nil
}
