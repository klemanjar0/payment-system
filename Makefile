DOCKER_COMPOSE_DIR = deployments/docker

.PHONY: proto sqlc sqlc-user sqlc-account sqlc-transaction \
	build run-local \
	run-user run-account run-transaction \
	build-user build-account build-transaction \
	migrate-up test

# --- Proto ---
proto:
	./scripts/generate.sh

# --- SQLC ---
sqlc:
	./scripts/sqlc-generate.sh

sqlc-user:
	./scripts/sqlc-generate.sh user

sqlc-account:
	./scripts/sqlc-generate.sh account

sqlc-transaction:
	./scripts/sqlc-generate.sh transaction

# --- Docker: all services ---
build:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.yaml build

run-local:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.yaml up

run-local-d:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.yaml up -d

down:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.yaml down

# --- Docker: individual services ---
build-user:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.user.yaml build

run-user:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.user.yaml up

down-user:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.user.yaml down

build-account:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.account.yaml build

run-account:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.account.yaml up

down-account:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.account.yaml down

build-transaction:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.transaction.yaml build

run-transaction:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.transaction.yaml up

down-transaction:
	docker compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.transaction.yaml down

# --- Other ---
migrate-up:
	./scripts/migrate.sh up

test:
	go test ./...
