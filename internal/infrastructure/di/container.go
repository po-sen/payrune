package di

import (
	"os"

	httpcontroller "payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/outbound/bitcoin"
	"payrune/internal/adapters/outbound/system"
	"payrune/internal/application/use_cases"
	"payrune/internal/domain/value_objects"
)

type Container struct {
	HealthController       *httpcontroller.HealthController
	ChainAddressController *httpcontroller.ChainAddressController
}

func NewContainer() *Container {
	clock := system.NewClock()
	healthUseCase := use_cases.NewCheckHealthUseCase(clock)
	healthController := httpcontroller.NewHealthController(healthUseCase)

	bitcoinDeriver := bitcoin.NewHDXPubAddressDeriver(
		bitcoin.NewLegacyAddressEncoder(),
		bitcoin.NewSegwitAddressEncoder(),
		bitcoin.NewNativeSegwitAddressEncoder(),
		bitcoin.NewTaprootAddressEncoder(),
	)
	addressPolicyCatalog := use_cases.NewAddressPolicyCatalog([]use_cases.AddressPolicyConfig{
		{
			AddressPolicyID: "bitcoin-mainnet-legacy",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeLegacy,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            os.Getenv("BITCOIN_MAINNET_LEGACY_XPUB"),
		},
		{
			AddressPolicyID: "bitcoin-mainnet-segwit",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeSegwit,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            os.Getenv("BITCOIN_MAINNET_SEGWIT_XPUB"),
		},
		{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeNativeSegwit,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            os.Getenv("BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"),
		},
		{
			AddressPolicyID: "bitcoin-mainnet-taproot",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkMainnet,
			Scheme:          value_objects.BitcoinAddressSchemeTaproot,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            os.Getenv("BITCOIN_MAINNET_TAPROOT_XPUB"),
		},
		{
			AddressPolicyID: "bitcoin-testnet4-legacy",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkTestnet4,
			Scheme:          value_objects.BitcoinAddressSchemeLegacy,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            os.Getenv("BITCOIN_TESTNET4_LEGACY_XPUB"),
		},
		{
			AddressPolicyID: "bitcoin-testnet4-segwit",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkTestnet4,
			Scheme:          value_objects.BitcoinAddressSchemeSegwit,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            os.Getenv("BITCOIN_TESTNET4_SEGWIT_XPUB"),
		},
		{
			AddressPolicyID: "bitcoin-testnet4-native-segwit",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkTestnet4,
			Scheme:          value_objects.BitcoinAddressSchemeNativeSegwit,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            os.Getenv("BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"),
		},
		{
			AddressPolicyID: "bitcoin-testnet4-taproot",
			Chain:           value_objects.ChainBitcoin,
			Network:         value_objects.BitcoinNetworkTestnet4,
			Scheme:          value_objects.BitcoinAddressSchemeTaproot,
			MinorUnit:       "satoshi",
			Decimals:        8,
			XPub:            os.Getenv("BITCOIN_TESTNET4_TAPROOT_XPUB"),
		},
	})
	listAddressPoliciesUseCase := use_cases.NewListAddressPoliciesUseCase(addressPolicyCatalog)
	generateAddressUseCase := use_cases.NewGenerateAddressUseCase(bitcoinDeriver, addressPolicyCatalog)
	chainAddressController := httpcontroller.NewChainAddressController(listAddressPoliciesUseCase, generateAddressUseCase)

	return &Container{
		HealthController:       healthController,
		ChainAddressController: chainAddressController,
	}
}
