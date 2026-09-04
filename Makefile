.PHONY: run test up down

run:
	go run ./cmd/server

test:
	go test ./...

up:
	docker-compose up -d

down:
	docker-compose down
