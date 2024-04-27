build:
	@go build -o bin/the-button-server

run: build
	@./bin/the-button-server

test:
	@go test -v ./...
