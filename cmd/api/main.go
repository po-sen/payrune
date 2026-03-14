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

	if err := bootstrap.RunAPI(ctx, ":8080"); err != nil {
		log.Fatalf("service exited with error: %v", err)
	}
}
