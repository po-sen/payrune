package valueobjects

import "strings"

type NetworkID string

func ParseNetworkID(raw string) (NetworkID, bool) {
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

	return NetworkID(normalized), true
}
