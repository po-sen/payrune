package value_objects

import "strings"

type Chain string

const (
	ChainBitcoin Chain = "bitcoin"
)

func ParseChain(raw string) (Chain, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(ChainBitcoin):
		return ChainBitcoin, true
	default:
		return "", false
	}
}
