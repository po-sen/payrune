package value_objects

type Chain string

const (
	ChainBitcoin Chain = "bitcoin"
)

func ParseChain(raw string) (Chain, bool) {
	chainID, ok := ParseChainID(raw)
	if !ok {
		return "", false
	}

	switch chainID {
	case ChainIDBitcoin:
		return ChainBitcoin, true
	default:
		return "", false
	}
}
