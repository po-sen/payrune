package valueobjects

import "strings"

type BitcoinAddressScheme string

const (
	BitcoinAddressSchemeLegacy       BitcoinAddressScheme = "legacy"
	BitcoinAddressSchemeSegwit       BitcoinAddressScheme = "segwit"
	BitcoinAddressSchemeNativeSegwit BitcoinAddressScheme = "nativeSegwit"
	BitcoinAddressSchemeTaproot      BitcoinAddressScheme = "taproot"
)

var bitcoinAddressSchemes = map[string]BitcoinAddressScheme{
	"legacy":       BitcoinAddressSchemeLegacy,
	"segwit":       BitcoinAddressSchemeSegwit,
	"nativesegwit": BitcoinAddressSchemeNativeSegwit,
	"taproot":      BitcoinAddressSchemeTaproot,
}

func ParseBitcoinAddressScheme(raw string) (BitcoinAddressScheme, bool) {
	scheme, ok := bitcoinAddressSchemes[strings.ToLower(strings.TrimSpace(raw))]
	return scheme, ok
}
