COMPOSE := docker compose \
	--env-file deployments/compose/compose.test.env \
	-f deployments/compose/compose.yaml \
	-f deployments/compose/compose.test.yaml

.PHONY: up down

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down
