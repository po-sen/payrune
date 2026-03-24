package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"payrune/internal/bootstrap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := bootstrap.LoadReceiptWebhookDispatcherConfigFromEnv()
	if err != nil {
		log.Fatalf("invalid webhook dispatcher config: %v", err)
	}

	if err := bootstrap.RunReceiptWebhookDispatcher(ctx, config); err != nil {
		log.Fatalf("webhook dispatcher exited with error: %v", err)
	}
}
