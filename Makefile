.PHONY: up down logs test

up:
	docker-compose up -d --build
down:
	docker-compose down -v
logs:
	docker-compose logs -f --tail=100
test:
	go test ./...