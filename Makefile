.PHONY: build run clean test

build:
	@mkdir -p bin
	@go build -o bin/sonarqube-mcp .

run: build
	@bin/sonarqube-mcp

clean:
	@rm -f bin/sonarqube-mcp

test:
	@go test ./...
