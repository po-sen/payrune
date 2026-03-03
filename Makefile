COMPOSE_FILE := deployments/compose/compose.yaml
COMPOSE := docker compose -f $(COMPOSE_FILE)

.PHONY: up down

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down
