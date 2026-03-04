.PHONY: build test lint clean frontend

# Build the frontend and Go binary
build: frontend
	go build -o bin/seance ./cmd/seance

# Build just the frontend
frontend:
	cd web/frontend && npm ci --silent && npx vite build

# Run tests
test:
	go test ./...

# Run linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/ web/dist/

# Run the server locally
run: build
	./bin/seance
