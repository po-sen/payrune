package valueobjects

type SupportedChain string

const (
	SupportedChainBitcoin SupportedChain = "bitcoin"
)

func ParseSupportedChain(raw string) (SupportedChain, bool) {
	chainID, ok := ParseChainID(raw)
	if !ok {
		return "", false
	}

	switch chainID {
	case ChainIDBitcoin:
		return SupportedChainBitcoin, true
	default:
		return "", false
	}
}
