package out

import "payrune/internal/domain/value_objects"

type BitcoinAddressDeriver interface {
	DeriveAddress(
		network value_objects.BitcoinNetwork,
		scheme value_objects.BitcoinAddressScheme,
		xpub string,
		index uint32,
	) (string, error)
}
