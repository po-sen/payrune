package bitcoin

import (
	"strings"
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

func parseAddressScheme(raw string) (addressScheme, bool) {
	scheme, ok := bitcoinAddressSchemes[strings.ToLower(strings.TrimSpace(raw))]
	return scheme, ok
}
