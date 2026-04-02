package ethereum

import (
	"encoding/json"
	"fmt"
	"strings"

	"payrune/internal/domain/policies"
	ethereumcreate2assets "payrune/internal/infrastructure/ethereumcreate2assets"
)

const create2SweepMaterialVersion = 1

type sweepMaterial struct {
	MaterialType     string `json:"material_type"`
	MaterialVersion  int    `json:"material_version"`
	Chain            string `json:"chain"`
	Network          string `json:"network"`
	Address          string `json:"address"`
	PredictedAddress string `json:"predicted_address"`
	FactoryAddress   string `json:"factory_address"`
	CollectorAddress string `json:"collector_address"`
	Create2Salt      string `json:"create2_salt"`
	InitCodeHex      string `json:"init_code_hex"`
	InitCodeHash     string `json:"init_code_hash"`
}

func buildSweepMaterialJSON(
	policy policies.AddressIssuancePolicy,
	address string,
	create2Salt string,
) (string, error) {
	policy = policy.Normalize()

	sourceRef, err := parseCreate2SourceRef(policy.IssuanceConfig.AddressSpaceRef)
	if err != nil {
		return "", err
	}

	metadata, ok := ethereumcreate2assets.LookupDeploymentMetadata(string(policy.Network))
	if !ok {
		return "", fmt.Errorf("ethereum create2 metadata not found for network: %s", policy.Network)
	}

	initCodeHex, ok := metadata.Receiver.InitCodeHex(sourceRef.collector)
	if !ok {
		return "", fmt.Errorf("ethereum create2 init code is unavailable for collector: %s", sourceRef.collector)
	}

	raw, err := json.Marshal(sweepMaterial{
		MaterialType:     "ethereum_create2",
		MaterialVersion:  create2SweepMaterialVersion,
		Chain:            string(policy.Chain),
		Network:          string(policy.Network),
		Address:          strings.TrimSpace(address),
		PredictedAddress: strings.TrimSpace(address),
		FactoryAddress:   sourceRef.factoryAddress,
		CollectorAddress: sourceRef.collector,
		Create2Salt:      strings.TrimSpace(create2Salt),
		InitCodeHex:      initCodeHex,
		InitCodeHash:     sourceRef.initCodeHash,
	})
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
