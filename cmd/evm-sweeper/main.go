package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"payrune/internal/bootstrap"
	"payrune/internal/domain/valueobjects"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := parseEVMSweeperConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("invalid evm sweeper config: %v", err)
	}

	if err := bootstrap.RunEVMSweeper(ctx, config); err != nil {
		log.Fatalf("evm sweeper exited with error: %v", err)
	}
}

func parseEVMSweeperConfig(args []string) (bootstrap.EVMSweeperConfig, error) {
	flagSet := flag.NewFlagSet("evm-sweeper", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)

	var (
		networkRaw           string
		assetCodeRaw         string
		paymentAddressIDsRaw string
		beforeIssuedAtRaw    string
		batchSize            int
		dryRun               bool
	)

	flagSet.StringVar(&networkRaw, "network", "", "filter by network")
	flagSet.StringVar(&assetCodeRaw, "asset-code", "", "filter by asset code")
	flagSet.StringVar(&paymentAddressIDsRaw, "payment-address-ids", "", "comma-separated payment address ids")
	flagSet.StringVar(&beforeIssuedAtRaw, "before-issued-at", "", "RFC3339 timestamp upper bound for issuedAt")
	flagSet.IntVar(&batchSize, "batch-size", 0, "maximum number of rows to process")
	flagSet.BoolVar(&dryRun, "dry-run", true, "log selected filters without executing sweep")

	if err := flagSet.Parse(args); err != nil {
		return bootstrap.EVMSweeperConfig{}, err
	}

	network, err := parseOptionalNetwork(networkRaw)
	if err != nil {
		return bootstrap.EVMSweeperConfig{}, err
	}
	assetCode, err := parseOptionalAssetCode(assetCodeRaw)
	if err != nil {
		return bootstrap.EVMSweeperConfig{}, err
	}
	paymentAddressIDs, err := parseOptionalPaymentAddressIDs(paymentAddressIDsRaw)
	if err != nil {
		return bootstrap.EVMSweeperConfig{}, err
	}
	beforeIssuedAt, err := parseOptionalRFC3339(beforeIssuedAtRaw)
	if err != nil {
		return bootstrap.EVMSweeperConfig{}, err
	}
	if batchSize < 0 {
		return bootstrap.EVMSweeperConfig{}, errorsf("batch-size must be greater than or equal to zero")
	}

	return bootstrap.EVMSweeperConfig{
		Network:           network,
		AssetCode:         assetCode,
		PaymentAddressIDs: paymentAddressIDs,
		BeforeIssuedAt:    beforeIssuedAt,
		BatchSize:         batchSize,
		DryRun:            dryRun,
	}, nil
}

func parseOptionalNetwork(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	network, ok := valueobjects.ParseNetworkID(trimmed)
	if !ok {
		return "", errorsf("network is invalid")
	}
	return string(network), nil
}

func parseOptionalAssetCode(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", nil
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
		return "", errorsf("asset-code is invalid")
	}
	return normalized, nil
}

func parseOptionalPaymentAddressIDs(raw string) ([]int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	parts := strings.Split(trimmed, ",")
	values := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, errorsf("payment-address-ids is invalid")
		}
		value, err := strconv.ParseInt(part, 10, 64)
		if err != nil || value <= 0 {
			return nil, errorsf("payment-address-ids is invalid")
		}
		values = append(values, value)
	}
	return values, nil
}

func parseOptionalRFC3339(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, nil
	}
	value, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return time.Time{}, errorsf("before-issued-at must be RFC3339: %v", err)
	}
	return value.UTC(), nil
}

func errorsf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
