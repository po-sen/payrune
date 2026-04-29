package ethereum

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	outport "payrune/internal/application/ports/outbound"
	ethereumcreate2assets "payrune/internal/infrastructure/ethereumcreate2assets"
)

const zeroEthereumAddress = "0x0000000000000000000000000000000000000000"

type EthereumAddressIssuanceReadinessChecker struct {
	clients map[string]*ethereumRPCClient
}

func NewAddressIssuanceReadinessChecker(
	configs map[string]*EthereumRPCObserverConfig,
) (*EthereumAddressIssuanceReadinessChecker, error) {
	clients := make(map[string]*ethereumRPCClient, len(configs))
	for rawNetwork, config := range configs {
		network, ok := outport.NormalizeNetworkID(rawNetwork)
		if !ok {
			return nil, fmt.Errorf("ethereum network is invalid: %s", rawNetwork)
		}

		client, err := newEthereumRPCClient(config)
		if err != nil {
			return nil, fmt.Errorf("configure %s ethereum rpc client: %w", network, err)
		}
		if client == nil {
			continue
		}

		clients[network] = client
	}

	return &EthereumAddressIssuanceReadinessChecker{clients: clients}, nil
}

func (c *EthereumAddressIssuanceReadinessChecker) CheckIssuanceReadiness(
	ctx context.Context,
	policy outport.AddressIssuancePolicyRecord,
) error {
	if policy.Chain != outport.SupportedChainEthereum || !policy.Enabled {
		return nil
	}

	client, ok := c.clients[policy.Network]
	if !ok || client == nil {
		return c.policyUnavailableError(
			policy,
			"ethereum rpc client is not configured for network %s",
			policy.Network,
		)
	}

	if err := c.checkFactory(ctx, policy, client); err != nil {
		return err
	}

	assetReference := strings.TrimSpace(policy.AssetReference)
	if assetReference == "" {
		return nil
	}

	return c.checkTokenContract(ctx, policy, client, assetReference)
}

func (c *EthereumAddressIssuanceReadinessChecker) checkFactory(
	ctx context.Context,
	policy outport.AddressIssuancePolicyRecord,
	client *ethereumRPCClient,
) error {
	metadata, ok := ethereumcreate2assets.LookupDeploymentMetadata(policy.Network)
	if !ok {
		return c.policyUnavailableError(policy, "ethereum create2 metadata is missing")
	}

	factoryAddress, err := NormalizeEthereumAddress(metadata.FactoryAddress, "factory address")
	if err != nil {
		return c.policyUnavailableError(policy, "factory address is invalid: %v", err)
	}

	rawCode, err := client.fetchCode(ctx, factoryAddress)
	if err != nil {
		return c.policyUnavailableError(policy, "factory code check failed: %v", err)
	}

	actualHash, deployed, err := runtimeCodeHashHex(rawCode)
	if err != nil {
		return c.policyUnavailableError(policy, "factory code is invalid: %v", err)
	}
	if !deployed {
		return c.policyUnavailableError(policy, "factory contract is not deployed")
	}

	expectedHash, ok := ethereumcreate2assets.ExpectedFactoryRuntimeCodeHashHex()
	if !ok {
		return c.policyUnavailableError(policy, "expected factory runtime hash is unavailable")
	}
	if actualHash != expectedHash {
		return c.policyUnavailableError(
			policy,
			"factory runtime hash mismatch: got %s want %s",
			actualHash,
			expectedHash,
		)
	}

	return nil
}

func (c *EthereumAddressIssuanceReadinessChecker) checkTokenContract(
	ctx context.Context,
	policy outport.AddressIssuancePolicyRecord,
	client *ethereumRPCClient,
	assetReference string,
) error {
	tokenAddress, err := NormalizeEthereumAddress(assetReference, "asset reference")
	if err != nil {
		return c.policyUnavailableError(policy, "asset reference is invalid: %v", err)
	}

	rawCode, err := client.fetchCode(ctx, tokenAddress)
	if err != nil {
		return c.policyUnavailableError(policy, "token code check failed: %v", err)
	}
	if _, deployed, err := runtimeCodeHashHex(rawCode); err != nil {
		return c.policyUnavailableError(policy, "token code is invalid: %v", err)
	} else if !deployed {
		return c.policyUnavailableError(policy, "token contract is not deployed")
	}

	balanceCallData, err := encodeERC20BalanceOfCall(zeroEthereumAddress)
	if err != nil {
		return c.policyUnavailableError(policy, "zero-address balance call is invalid: %v", err)
	}
	rawBalance, err := client.callContractAtLatest(ctx, tokenAddress, balanceCallData)
	if err != nil {
		return c.policyUnavailableError(policy, "token balanceOf call failed: %v", err)
	}
	if _, err := parseEthereumHexQuantity(rawBalance, "token balance"); err != nil {
		return c.policyUnavailableError(policy, "token balanceOf response is invalid: %v", err)
	}

	decimals, err := client.fetchERC20Decimals(ctx, tokenAddress)
	if err != nil {
		return c.policyUnavailableError(policy, "token decimals call failed: %v", err)
	}
	if decimals != policy.Decimals {
		return c.policyUnavailableError(
			policy,
			"token decimals mismatch: got %d want %d",
			decimals,
			policy.Decimals,
		)
	}

	return nil
}

func (c *EthereumAddressIssuanceReadinessChecker) policyUnavailableError(
	policy outport.AddressIssuancePolicyRecord,
	format string,
	args ...any,
) error {
	return fmt.Errorf(
		"ethereum address issuance policy is unavailable: policy=%s chain=%s network=%s: %s",
		policy.AddressPolicyID,
		policy.Chain,
		policy.Network,
		fmt.Sprintf(format, args...),
	)
}

func runtimeCodeHashHex(rawCode string) (string, bool, error) {
	codeBytes, err := parseEthereumHexBytes(rawCode, "runtime code")
	if err != nil {
		return "", false, err
	}
	if len(codeBytes) == 0 {
		return "", false, nil
	}

	hash := keccak256Hash(codeBytes)
	return "0x" + hex.EncodeToString(hash[:]), true, nil
}
