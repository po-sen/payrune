.DEFAULT_GOAL := help

COMPOSE_FILE := deployments/compose/compose.yaml
COMPOSE_ENV := deployments/compose/compose.env
COMPOSE_DEV_ENV := deployments/compose/compose.dev.env

.PHONY: \
	help \
	up \
	down \
	config \
	up-mainnet \
	down-mainnet \
	config-mainnet \
	cf-up \
	cf-down

help:
	@printf "%s\n" \
		"up             start base stack plus development-profile services" \
		"down           stop base stack plus development-profile services" \
		"config         render base stack plus development-profile services" \
		"up-mainnet     start base stack only" \
		"down-mainnet   stop base stack only" \
		"config-mainnet render base-stack-only compose config" \
		"cf-up          migrate and deploy Cloudflare workers" \
		"cf-down        delete Cloudflare workers"

up down config: $(COMPOSE_DEV_ENV)

up-mainnet down-mainnet config-mainnet: $(COMPOSE_ENV)

up:
	docker compose --env-file $(COMPOSE_DEV_ENV) --profile development -f $(COMPOSE_FILE) up -d --build

down:
	docker compose --env-file $(COMPOSE_DEV_ENV) --profile development -f $(COMPOSE_FILE) down

config:
	docker compose --env-file $(COMPOSE_DEV_ENV) --profile development -f $(COMPOSE_FILE) config

up-mainnet:
	docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) up -d --build

down-mainnet:
	docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) down

config-mainnet:
	docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) config

cf-up:
	./scripts/cf-cloudflare-migrate.sh
	./scripts/cf-receipt-webhook-mock-worker-deploy.sh
	./scripts/cf-payrune-worker-deploy.sh

cf-down:
	./scripts/cf-payrune-worker-delete.sh
	./scripts/cf-receipt-webhook-mock-worker-delete.sh
