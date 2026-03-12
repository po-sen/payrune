package valueobjects

import "strings"

type ChainID string

const (
	ChainIDBitcoin ChainID = "bitcoin"
	maxIDLength    int     = 64
)

func ParseChainID(raw string) (ChainID, bool) {
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

	return ChainID(normalized), true
}
