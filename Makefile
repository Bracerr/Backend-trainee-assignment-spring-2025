.PHONY: test test-unit test-integration

test:
	docker-compose exec -T app go test ./...

test-unit:
	docker-compose exec -T app go test ./src/tests/unit/...

test-integration:
	docker-compose exec -T app go test ./src/tests/integration/... 