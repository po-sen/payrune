package valueobjects

import "strings"

type AddressScheme string

const (
	AddressSchemeLegacy       AddressScheme = "legacy"
	AddressSchemeSegwit       AddressScheme = "segwit"
	AddressSchemeNativeSegwit AddressScheme = "nativeSegwit"
	AddressSchemeTaproot      AddressScheme = "taproot"
	AddressSchemeCreate2      AddressScheme = "create2"
)

var addressSchemes = map[string]AddressScheme{
	"legacy":       AddressSchemeLegacy,
	"segwit":       AddressSchemeSegwit,
	"nativesegwit": AddressSchemeNativeSegwit,
	"taproot":      AddressSchemeTaproot,
	"create2":      AddressSchemeCreate2,
}

func ParseAddressScheme(raw string) (AddressScheme, bool) {
	scheme, ok := addressSchemes[strings.ToLower(strings.TrimSpace(raw))]
	return scheme, ok
}

func (s AddressScheme) Normalize() AddressScheme {
	if normalized, ok := ParseAddressScheme(string(s)); ok {
		return normalized
	}
	return AddressScheme(strings.TrimSpace(string(s)))
}

func (s AddressScheme) IsZero() bool {
	return strings.TrimSpace(string(s)) == ""
}

func (s AddressScheme) IsCreate2() bool {
	return s.Normalize() == AddressSchemeCreate2
}
