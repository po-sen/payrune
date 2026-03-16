package valueobjects

type SupportedChain string

const (
	SupportedChainBitcoin  SupportedChain = "bitcoin"
	SupportedChainEthereum SupportedChain = "ethereum"
)

func ParseSupportedChain(raw string) (SupportedChain, bool) {
	chainID, ok := ParseChainID(raw)
	if !ok {
		return "", false
	}

	switch chainID {
	case ChainIDBitcoin:
		return SupportedChainBitcoin, true
	case ChainIDEthereum:
		return SupportedChainEthereum, true
	default:
		return "", false
	}
}
