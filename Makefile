COMPOSE_FILE := deployments/compose/compose.yaml
COMPOSE_OVERRIDE ?= \
	deployments/compose/compose.bitcoin.testnet4.yaml \
	deployments/compose/compose.test.yaml
COMPOSE_ENV_FILE ?= deployments/compose/compose.test.env
comma := ,
COMPOSE_OVERRIDE_LIST := $(strip $(subst $(comma), ,$(COMPOSE_OVERRIDE)))
COMPOSE_FILES := -f $(COMPOSE_FILE) $(foreach file,$(COMPOSE_OVERRIDE_LIST),-f $(file))
COMPOSE_ENV_ARG := $(if $(strip $(COMPOSE_ENV_FILE)),--env-file $(COMPOSE_ENV_FILE),)
COMPOSE := docker compose $(COMPOSE_ENV_ARG) $(COMPOSE_FILES)

.PHONY: up down

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down
