package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"payrune/internal/bootstrap"
	"payrune/internal/domain/value_objects"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := loadPollerConfigFromEnv()
	if err != nil {
		log.Fatalf("invalid poller config: %v", err)
	}

	if err := bootstrap.RunPoller(ctx, config); err != nil {
		log.Fatalf("poller exited with error: %v", err)
	}
}

func loadPollerConfigFromEnv() (bootstrap.PollerConfig, error) {
	interval, err := parseDurationEnv("POLL_INTERVAL")
	if err != nil {
		return bootstrap.PollerConfig{}, err
	}

	claimTTL, err := parseDurationEnv("POLL_CLAIM_TTL")
	if err != nil {
		return bootstrap.PollerConfig{}, err
	}

	batchSize, err := parseIntEnv("POLL_BATCH_SIZE")
	if err != nil {
		return bootstrap.PollerConfig{}, err
	}
	chain, err := parseChainEnv("POLL_CHAIN")
	if err != nil {
		return bootstrap.PollerConfig{}, err
	}
	network, err := parseNetworkEnv("POLL_NETWORK")
	if err != nil {
		return bootstrap.PollerConfig{}, err
	}
	if network != "" && chain == "" {
		return bootstrap.PollerConfig{}, fmt.Errorf("POLL_CHAIN is required when POLL_NETWORK is set")
	}

	return bootstrap.PollerConfig{
		Interval:  interval,
		BatchSize: batchSize,
		ClaimTTL:  claimTTL,
		Chain:     chain,
		Network:   network,
	}, nil
}

func parseDurationEnv(key string) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func parseIntEnv(key string) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func parseChainEnv(key string) (string, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return "", nil
	}

	chain, ok := value_objects.ParseChainID(raw)
	if !ok {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return string(chain), nil
}

func parseNetworkEnv(key string) (string, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return "", nil
	}

	network, ok := value_objects.ParseNetworkID(raw)
	if !ok {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return string(network), nil
}
