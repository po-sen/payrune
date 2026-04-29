package bitcoin

import (
	"strings"
)

type network string

const (
	networkMainnet  network = "mainnet"
	networkTestnet4 network = "testnet4"
)

var bitcoinNetworks = map[string]network{
	"mainnet":  networkMainnet,
	"testnet4": networkTestnet4,
}

func parseNetwork(raw string) (network, bool) {
	network, ok := bitcoinNetworks[strings.ToLower(strings.TrimSpace(raw))]
	return network, ok
}

func (n network) NetworkID() string {
	return string(n)
}
