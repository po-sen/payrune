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

	config, err := bootstrap.LoadPollerConfigFromEnv()
	if err != nil {
		log.Fatalf("invalid poller config: %v", err)
	}

	if err := bootstrap.RunPoller(ctx, config); err != nil {
		log.Fatalf("poller exited with error: %v", err)
	}
}
