package main

import (
	"log"

	"payrune/internal/bootstrap"
)

func main() {
	config := bootstrap.LoadFakeWebhookReceiverConfigFromEnv()
	if err := bootstrap.RunFakeWebhookReceiver(config); err != nil {
		log.Fatalf("fake webhook receiver exited with error: %v", err)
	}
}
