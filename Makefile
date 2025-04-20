.PHONY: test test-unit test-integration

test:
	docker-compose exec -T app go test ./...

test-unit:
	docker-compose exec -T app go test ./...

test-integration:
	docker-compose exec -T app go test -tags=integration ./src/tests/integration/... 