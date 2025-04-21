up-daemon:
	docker-compose up -d

up:
	docker-compose up

down:
	docker-compose down

test-integration: check-docker
	docker-compose exec -T app go test -tags=integration ./src/tests/integration/...

test-unit: check-docker
	docker-compose exec -T app go test ./...

check-docker:
	@docker-compose ps app | grep Up || (echo "Docker containers not running. Run 'make up-daemon' first" && exit 1)