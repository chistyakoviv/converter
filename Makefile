cli-test:
	docker compose run --rm go-cli CONFIG_PATH=config/local.yml go run cmd/converter/main.go