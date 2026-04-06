package valueobjects

import "strings"

type AddressPolicyID string

const (
	AddressPolicyIDBitcoinMainnetLegacy        AddressPolicyID = "bitcoin-mainnet-legacy"
	AddressPolicyIDBitcoinMainnetSegwit        AddressPolicyID = "bitcoin-mainnet-segwit"
	AddressPolicyIDBitcoinMainnetNativeSegwit  AddressPolicyID = "bitcoin-mainnet-native-segwit"
	AddressPolicyIDBitcoinMainnetTaproot       AddressPolicyID = "bitcoin-mainnet-taproot"
	AddressPolicyIDBitcoinTestnet4Legacy       AddressPolicyID = "bitcoin-testnet4-legacy"
	AddressPolicyIDBitcoinTestnet4Segwit       AddressPolicyID = "bitcoin-testnet4-segwit"
	AddressPolicyIDBitcoinTestnet4NativeSegwit AddressPolicyID = "bitcoin-testnet4-native-segwit"
	AddressPolicyIDBitcoinTestnet4Taproot      AddressPolicyID = "bitcoin-testnet4-taproot"
	AddressPolicyIDEthereumMainnetCreate2      AddressPolicyID = "ethereum-mainnet-create2"
	AddressPolicyIDEthereumSepoliaCreate2      AddressPolicyID = "ethereum-sepolia-create2"
	AddressPolicyIDEthereumMainnetUSDTCreate2  AddressPolicyID = "ethereum-mainnet-usdt-create2"
	AddressPolicyIDEthereumSepoliaUSDTCreate2  AddressPolicyID = "ethereum-sepolia-usdt-create2"
)

func NewAddressPolicyID(raw string) (AddressPolicyID, error) {
	normalized, ok := parseAddressPolicyID(raw)
	if !ok {
		return "", ErrAddressPolicyIDInvalid
	}
	return normalized, nil
}

func parseAddressPolicyID(raw string) (AddressPolicyID, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" || len(normalized) > maxIDLength {
		return "", false
	}

	for i := 0; i < len(normalized); i++ {
		char := normalized[i]
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '_' || char == '-' {
			continue
		}
		return "", false
	}

	return AddressPolicyID(normalized), true
}

func (id AddressPolicyID) Normalize() AddressPolicyID {
	normalized, ok := parseAddressPolicyID(string(id))
	if !ok {
		return ""
	}
	return normalized
}

func (id AddressPolicyID) IsZero() bool {
	return id.Normalize() == ""
}

func EthereumCreate2AddressPolicyID(network NetworkID) AddressPolicyID {
	normalized, ok := ParseNetworkID(string(network))
	if !ok {
		return ""
	}

	switch normalized {
	case NetworkIDMainnet:
		return AddressPolicyIDEthereumMainnetCreate2
	case NetworkIDSepolia:
		return AddressPolicyIDEthereumSepoliaCreate2
	default:
		return AddressPolicyID("ethereum-" + string(normalized) + "-create2")
	}
}

func EthereumUSDTCreate2AddressPolicyID(network NetworkID) AddressPolicyID {
	normalized, ok := ParseNetworkID(string(network))
	if !ok {
		return ""
	}

	switch normalized {
	case NetworkIDMainnet:
		return AddressPolicyIDEthereumMainnetUSDTCreate2
	case NetworkIDSepolia:
		return AddressPolicyIDEthereumSepoliaUSDTCreate2
	default:
		return AddressPolicyID("ethereum-" + string(normalized) + "-usdt-create2")
	}
}
