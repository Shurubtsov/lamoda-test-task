up:
	docker-compose up

down:
	docker-compose down -v

build:
	go mod download && go mod tidy

.PHONY: up down build