.PHONY: run
run:
	@set -a; . ./.env; set +a; go run ./cmd/api -config config/config.yaml