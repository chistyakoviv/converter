# Project commands
cli-test:
	docker compose run --rm go-cli sh -c "CONFIG_PATH=config/local.yml go run cmd/converter/main.go"

migrate-status:
	docker compose run --rm go-cli make converter-migrate-status

migrate-up:
	docker compose run --rm go-cli make converter-migrate-up

migrate-down:
	docker compose run --rm go-cli make converter-migrate-down

migration:
	docker compose run --rm go-cli goose create ${MIGRATION_NAME} sql

.PHONY: mocks
mocks:
	docker compose run --rm go-cli mockery

tests:
	docker compose run --rm converter go test -v ./...

db-connect:
	docker compose exec converter-pg psql postgres://app:secret@converter-pg/app

db-purge:
	docker compose exec converter-pg sh -c "psql postgres://app:secret@converter-pg/app -t -c \"SELECT 'DROP TABLE \\\"' || tablename || '\\\" CASCADE;' FROM pg_tables WHERE schemaname = 'public'\" | psql postgres://app:secret@converter-pg/app"

wait-db:
	wait-for-it converter-pg:5432 -t 60

# DB commands
converter-migrate-status: wait-db
	goose postgres "${PG_DSN}" status -v

converter-migrate-up: wait-db
	goose postgres "${PG_DSN}" up -v

converter-migrate-down: wait-db
	goose postgres "${PG_DSN}" down -v
