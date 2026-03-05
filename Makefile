COMPOSE_FILE := deployments/compose/compose.yaml
COMPOSE_OVERRIDE ?= \
	deployments/compose/compose.bitcoin.mainnet.yaml \
	deployments/compose/compose.bitcoin.testnet4.yaml \
	deployments/compose/compose.test-env.yaml
comma := ,
COMPOSE_OVERRIDE_LIST := $(strip $(subst $(comma), ,$(COMPOSE_OVERRIDE)))
COMPOSE_FILES := -f $(COMPOSE_FILE) $(foreach file,$(COMPOSE_OVERRIDE_LIST),-f $(file))
COMPOSE := docker compose $(COMPOSE_FILES)

.PHONY: up down

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down
