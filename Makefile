DATABASE_URL ?= postgres://analytics:analytics@localhost:5432/analytics?sslmode=disable
ADDR ?= :8080

.PHONY: generate mocks test run run-detached down docker-up wait-pg

generate:
	go generate ./ent/...

mocks:
	mockery

wait-pg:
	@echo "Waiting for PostgreSQL..."
	@until docker compose exec -T postgres pg_isready -U analytics -d analytics >/dev/null 2>&1; do sleep 1; done

docker-up:
	docker compose up -d postgres

down:
	docker compose down -v

# Start Postgres + API in Docker (curl http://localhost:8080/health)
run:
	docker compose up --build

run-detached:
	docker compose up -d --build

# Tests on host against compose Postgres on localhost:5432
test: docker-up wait-pg
	DATABASE_URL=$(DATABASE_URL) go test ./...
