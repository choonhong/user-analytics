.PHONY: generate test run

generate:
	go generate ./ent/...

test:
	go test ./...

run:
	DATABASE_PATH=./data/analytics.db ADDR=:8080 go run ./cmd/server
