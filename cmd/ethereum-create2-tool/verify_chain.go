package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var defaultVerifyFundAmountWei = big.NewInt(1)

type verifyChainOutput struct {
	Network              string `json:"network"`
	RPCURL               string `json:"rpcURL"`
	OperatorSigner       string `json:"operatorSigner"`
	Collector            string `json:"collector"`
	FactoryAddress       string `json:"factoryAddress"`
	PredictedAddress     string `json:"predictedAddress"`
	SourceRef            string `json:"sourceRef"`
	AddressReference     string `json:"addressReference"`
	Salt                 string `json:"salt"`
	ReceiverArtifact     string `json:"receiverArtifact"`
	ReceiverCodeDeployed bool   `json:"receiverCodeDeployed"`
	ReceiverBalanceWei   string `json:"receiverBalanceWei"`
	FundAmountWei        string `json:"fundAmountWei"`
}

func runVerifyChain(args []string) error {
	paths, err := resolveToolPaths()
	if err != nil {
		return err
	}

	flagSet := newFlagSet("verify-chain")
	network := flagSet.String("network", envOrDefault("ETHEREUM_CREATE2_VERIFY_NETWORK", ""), "Ethereum network label")
	rpcURL := flagSet.String("rpc-url", envOrDefault("ETHEREUM_CREATE2_VERIFY_RPC_URL", ""), "Ethereum JSON-RPC URL")
	operatorPrivateKey := flagSet.String("operator-private-key", envOrDefault("ETHEREUM_CREATE2_VERIFY_OPERATOR_PRIVATE_KEY", ""), "operator signer private key")
	collectorAddress := flagSet.String("collector", envOrDefault("ETHEREUM_CREATE2_VERIFY_COLLECTOR_ADDRESS", ""), "collector address")
	factoryAddressFlag := flagSet.String("factory", envOrDefault("ETHEREUM_CREATE2_VERIFY_FACTORY_ADDRESS", ""), "existing CREATE2 factory address; if omitted a new factory is deployed")
	salt := flagSet.String("salt", envOrDefault("ETHEREUM_CREATE2_VERIFY_SALT", ""), "32-byte CREATE2 salt hex; if omitted a random salt is generated")
	fundAmountWeiRaw := flagSet.String("fund-amount-wei", envOrDefault("ETHEREUM_CREATE2_VERIFY_FUND_AMOUNT_WEI", defaultVerifyFundAmountWei.String()), "amount to pre-fund the predicted address in wei")
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*network) == "" {
		return fmt.Errorf("network is required")
	}
	if strings.TrimSpace(*rpcURL) == "" {
		return fmt.Errorf("rpc url is required")
	}
	if strings.TrimSpace(*operatorPrivateKey) == "" {
		return fmt.Errorf("operator private key is required")
	}
	if strings.TrimSpace(*collectorAddress) == "" {
		return fmt.Errorf("collector address is required")
	}
	normalizedSalt, err := normalizeOrGenerateSalt(*salt)
	if err != nil {
		return err
	}

	fundAmountWei, err := parsePositiveBigInt(*fundAmountWeiRaw)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client, err := ethclient.DialContext(ctx, *rpcURL)
	if err != nil {
		return err
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return err
	}

	operatorKey, err := crypto.HexToECDSA(strings.TrimPrefix(strings.TrimSpace(*operatorPrivateKey), "0x"))
	if err != nil {
		return err
	}
	operatorAddress := crypto.PubkeyToAddress(operatorKey.PublicKey)

	factoryArtifact, err := loadReceiverArtifact(paths.factoryArtifact)
	if err != nil {
		return err
	}
	receiverArtifact, err := loadReceiverArtifact(paths.receiverArtifact)
	if err != nil {
		return err
	}

	factoryABI, err := abi.JSON(bytes.NewReader(factoryArtifact.ABI))
	if err != nil {
		return err
	}

	factoryAddress, err := ensureFactoryAddress(
		ctx,
		client,
		chainID,
		operatorKey,
		factoryABI,
		factoryArtifact.CreationCodeHex,
		*factoryAddressFlag,
	)
	if err != nil {
		return err
	}

	prediction, err := predictFromArtifact(
		ctx,
		*network,
		factoryAddress.Hex(),
		*collectorAddress,
		paths.receiverArtifact,
		normalizedSalt,
	)
	if err != nil {
		return err
	}

	receiverAddress := common.HexToAddress(prediction.PredictedAddress)
	receiverCode, err := client.CodeAt(ctx, receiverAddress, nil)
	if err != nil {
		return err
	}
	if len(receiverCode) > 0 {
		return fmt.Errorf("receiver code already exists at %s; choose a fresh salt", receiverAddress.Hex())
	}

	if _, err := sendNativeETH(
		ctx,
		client,
		operatorKey,
		chainID,
		receiverAddress,
		new(big.Int).Set(fundAmountWei),
	); err != nil {
		return err
	}

	factoryContract := bind.NewBoundContract(factoryAddress, factoryABI, client, client, client)
	receiverDeployAuth, err := bind.NewKeyedTransactorWithChainID(operatorKey, chainID)
	if err != nil {
		return err
	}
	receiverDeployAuth.Context = ctx

	deployReceiverTx, err := factoryContract.Transact(
		receiverDeployAuth,
		"deploy",
		common.HexToHash(prediction.Salt),
		common.FromHex(prediction.InitCodeHex),
	)
	if err != nil {
		return err
	}
	if _, err := bind.WaitMined(ctx, client, deployReceiverTx); err != nil {
		return err
	}

	receiverCode, err = client.CodeAt(ctx, receiverAddress, nil)
	if err != nil {
		return err
	}
	receiverBalance, err := client.BalanceAt(ctx, receiverAddress, nil)
	if err != nil {
		return err
	}
	if len(receiverCode) == 0 {
		return fmt.Errorf("receiver code was not deployed at %s", prediction.PredictedAddress)
	}
	if receiverBalance.Cmp(fundAmountWei) != 0 {
		return fmt.Errorf("unexpected receiver balance: got %s want %s", receiverBalance.String(), fundAmountWei.String())
	}

	return writePrettyJSON(os.Stdout, verifyChainOutput{
		Network:              strings.TrimSpace(*network),
		RPCURL:               *rpcURL,
		OperatorSigner:       operatorAddress.Hex(),
		Collector:            common.HexToAddress(*collectorAddress).Hex(),
		FactoryAddress:       factoryAddress.Hex(),
		PredictedAddress:     receiverAddress.Hex(),
		SourceRef:            prediction.SourceRef,
		AddressReference:     prediction.AddressReference,
		Salt:                 prediction.Salt,
		ReceiverArtifact:     receiverArtifact.ContractName,
		ReceiverCodeDeployed: true,
		ReceiverBalanceWei:   receiverBalance.String(),
		FundAmountWei:        fundAmountWei.String(),
	})
}

func ensureFactoryAddress(
	ctx context.Context,
	client *ethclient.Client,
	chainID *big.Int,
	operatorKey *ecdsa.PrivateKey,
	factoryABI abi.ABI,
	factoryCreationCodeHex string,
	rawFactoryAddress string,
) (common.Address, error) {
	rawFactoryAddress = strings.TrimSpace(rawFactoryAddress)
	if rawFactoryAddress != "" {
		address := common.HexToAddress(rawFactoryAddress)
		code, err := client.CodeAt(ctx, address, nil)
		if err != nil {
			return common.Address{}, err
		}
		if len(code) == 0 {
			return common.Address{}, fmt.Errorf("factory address %s has no deployed code", address.Hex())
		}
		return address, nil
	}

	deployAuth, err := bind.NewKeyedTransactorWithChainID(operatorKey, chainID)
	if err != nil {
		return common.Address{}, err
	}
	deployAuth.Context = ctx

	factoryAddress, factoryTx, _, err := bind.DeployContract(
		deployAuth,
		factoryABI,
		common.FromHex(factoryCreationCodeHex),
		client,
	)
	if err != nil {
		return common.Address{}, err
	}
	if _, err := bind.WaitMined(ctx, client, factoryTx); err != nil {
		return common.Address{}, err
	}
	return factoryAddress, nil
}

func parsePositiveBigInt(raw string) (*big.Int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("fund amount is required")
	}

	value, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, fmt.Errorf("invalid big integer value %q", raw)
	}
	if value.Sign() <= 0 {
		return nil, fmt.Errorf("fund amount must be greater than zero")
	}
	return value, nil
}

func sendNativeETH(
	ctx context.Context,
	client *ethclient.Client,
	operatorKey *ecdsa.PrivateKey,
	chainID *big.Int,
	to common.Address,
	value *big.Int,
) (*types.Receipt, error) {
	from := crypto.PubkeyToAddress(operatorKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, err
	}

	tipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	feeCap := new(big.Int).Set(tipCap)
	if header.BaseFee != nil {
		feeCap = new(big.Int).Add(new(big.Int).Mul(header.BaseFee, big.NewInt(2)), tipCap)
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &to,
		Value:     value,
		Gas:       21_000,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
	})

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), operatorKey)
	if err != nil {
		return nil, err
	}
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return nil, err
	}
	return bind.WaitMined(ctx, client, signedTx)
}
