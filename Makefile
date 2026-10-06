include .env
export

MIGRATIONS_DIR=migrations

.PHONY: run migrate-create migrate-up migrate-down
run:
	@go run ./cmd/api -config config/config.yaml

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: is migration name set? Example: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@mkdir -p $(MIGRATIONS_DIR)
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up:
	@migrate -path $(MIGRATIONS_DIR) -database "postgres://$$DB_USER:$$DB_PASSWORD@localhost:5432/$$DB_NAME?sslmode=disable" up

migrate-down:
	@migrate -path $(MIGRATIONS_DIR) -database "postgres://$$DB_USER:$$DB_PASSWORD@localhost:5432/$$DB_NAME?sslmode=disable" down 1