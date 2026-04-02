package bitcoin

import (
	"strings"

	"payrune/internal/domain/valueobjects"
)

type addressScheme string

const (
	addressSchemeLegacy       addressScheme = "legacy"
	addressSchemeSegwit       addressScheme = "segwit"
	addressSchemeNativeSegwit addressScheme = "nativeSegwit"
	addressSchemeTaproot      addressScheme = "taproot"
)

var bitcoinAddressSchemes = map[string]addressScheme{
	"legacy":       addressSchemeLegacy,
	"segwit":       addressSchemeSegwit,
	"nativesegwit": addressSchemeNativeSegwit,
	"taproot":      addressSchemeTaproot,
}

func parseAddressScheme(raw valueobjects.AddressScheme) (addressScheme, bool) {
	scheme, ok := bitcoinAddressSchemes[strings.ToLower(strings.TrimSpace(string(raw.Normalize())))]
	return scheme, ok
}
