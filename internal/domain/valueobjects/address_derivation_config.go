package valueobjects

import "strings"

type AddressIssuanceConfig struct {
	AddressSourceRef       string
	AddressReferencePrefix string
}

func (c AddressIssuanceConfig) Normalize() AddressIssuanceConfig {
	c.AddressSourceRef = strings.TrimSpace(c.AddressSourceRef)
	c.AddressReferencePrefix = normalizeAddressReferencePrefix(c.AddressReferencePrefix)
	return c
}

func (c AddressIssuanceConfig) IsEnabled() bool {
	return strings.TrimSpace(c.AddressSourceRef) != ""
}

func normalizeAddressReferencePrefix(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "m/") {
		return ""
	}
	return trimmed
}
