.PHONY: proto build run-local

proto:
	./scripts/generate.sh

build:
	docker-compose -f deployments/docker/docker-compose.yaml build

run-local:
	docker-compose -f deployments/docker/docker-compose.yaml up

migrate-up:
	./scripts/migrate.sh up

test:
	go test ./...