package valueobjects

import "strings"

type BitcoinNetwork string

const (
	BitcoinNetworkMainnet  BitcoinNetwork = "mainnet"
	BitcoinNetworkTestnet4 BitcoinNetwork = "testnet4"
)

var bitcoinNetworks = map[string]BitcoinNetwork{
	"mainnet":  BitcoinNetworkMainnet,
	"testnet4": BitcoinNetworkTestnet4,
}

func ParseBitcoinNetwork(raw string) (BitcoinNetwork, bool) {
	network, ok := bitcoinNetworks[strings.ToLower(strings.TrimSpace(raw))]
	return network, ok
}
