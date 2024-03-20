.PHONY: all arrange tidy lint test

all: arrange tidy lint test

arrange:
	@echo "Arranging files..."
	@go fmt ./...
	@goarrange run -r

tidy:
	@echo "Tidying up..."
	@go mod tidy

lint:
	@echo "Linting files..."
	@go vet ./...
	@golangci-lint run ./... -E misspell,godot,whitespace

test:
	@echo "Running tests..."
	@go test ./... -cover
