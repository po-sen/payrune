COMPOSE := docker compose \
	--env-file deployments/compose/compose.test.env \
	-f deployments/compose/compose.yaml \
	-f deployments/compose/compose.test.yaml

.PHONY: \
	up \
	down \
	cf-migrate \
	cf-up \
	cf-down

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

cf-migrate:
	./scripts/cf-cloudflare-migrate.sh

cf-up:
	./scripts/cf-cloudflare-migrate.sh
	./scripts/cf-api-worker-deploy.sh
	./scripts/cf-poller-worker-deploy.sh mainnet
	./scripts/cf-poller-worker-deploy.sh testnet4
	./scripts/cf-receipt-webhook-mock-worker-deploy.sh
	./scripts/cf-webhook-dispatcher-worker-deploy.sh

cf-down:
	./scripts/cf-webhook-dispatcher-worker-delete.sh
	./scripts/cf-receipt-webhook-mock-worker-delete.sh
	./scripts/cf-poller-worker-delete.sh testnet4
	./scripts/cf-poller-worker-delete.sh mainnet
	./scripts/cf-api-worker-delete.sh
