.PHONY: build test install clean benchmark fmt lint dev

BINARY_NAME=claude-code-hooks
BUILD_DIR=./bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/claude-code-hooks
	mkdir -p team-loops/bin 2>/dev/null || true
	cp $(BUILD_DIR)/$(BINARY_NAME) team-loops/bin/

test:
	go test -v ./...

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

clean:
	rm -rf $(BUILD_DIR)
	rm -rf team-loops/bin

benchmark:
	go test -bench=. ./...

fmt:
	go fmt ./...

lint:
	go vet ./...

tidy:
	go mod tidy

# Development
dev: build
	./$(BUILD_DIR)/$(BINARY_NAME) --help
