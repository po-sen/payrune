package valueobjects

import "strings"

type AddressIssuanceConfig struct {
	AddressSpaceRef   string
	IssuanceRefPrefix string
}

func (c AddressIssuanceConfig) Normalize() AddressIssuanceConfig {
	c.AddressSpaceRef = strings.TrimSpace(c.AddressSpaceRef)
	c.IssuanceRefPrefix = normalizeIssuanceRefPrefix(c.IssuanceRefPrefix)
	return c
}

func (c AddressIssuanceConfig) IsEnabled() bool {
	return strings.TrimSpace(c.AddressSpaceRef) != ""
}

func normalizeIssuanceRefPrefix(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimSuffix(trimmed, "/")
	return trimmed
}
