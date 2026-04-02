package bitcoin

import (
	"strings"

	"payrune/internal/domain/valueobjects"
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

func parseNetwork(raw valueobjects.NetworkID) (network, bool) {
	network, ok := bitcoinNetworks[strings.ToLower(strings.TrimSpace(string(raw)))]
	return network, ok
}

func (n network) NetworkID() valueobjects.NetworkID {
	return valueobjects.NetworkID(n)
}
