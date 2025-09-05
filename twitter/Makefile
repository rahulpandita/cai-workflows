.PHONY: build run clean deps

# Build the application
build:
	go build -o twitter-fetcher main.go

# Run the application
run:
	go run main.go

# Install dependencies
deps:
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	rm -f twitter-fetcher

# Run with environment variables loaded
run-with-env:
	@if [ -f .env ]; then \
		echo "Loading environment variables..."; \
		export $$(cat .env | xargs) && go run main.go; \
	else \
		echo ".env file not found. Please copy .env.example to .env and configure it."; \
	fi

# Development setup
setup:
	@echo "Setting up Twitter Fetcher..."
	go mod tidy
	cp .env.example .env
	@echo "Setup complete! Please edit .env file with your Twitter API credentials."
