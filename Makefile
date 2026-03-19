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
	./scripts/cf-receipt-webhook-mock-worker-deploy.sh
	./scripts/cf-payrune-worker-deploy.sh

cf-down:
	./scripts/cf-payrune-worker-delete.sh
	./scripts/cf-receipt-webhook-mock-worker-delete.sh
