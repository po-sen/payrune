COMPOSE := docker compose \
	--env-file deployments/compose/compose.test.env \
	-f deployments/compose/compose.yaml \
	-f deployments/compose/compose.test.yaml

.PHONY: \
	up \
	down \
	cf-api-deploy \
	cf-api-delete

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

cf-api-deploy:
	./scripts/cf-payrune-api-deploy.sh

cf-api-delete:
	./scripts/cf-payrune-api-delete.sh
