.PHONY: generate mocks test run

generate:
	go generate ./ent/...

mocks:
	mockery

test:
	go test ./...

run:
	DATABASE_PATH=./data/analytics.db ADDR=:8080 go run ./cmd/server
