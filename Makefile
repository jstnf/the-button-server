build:
	@go build -o bin/the-button-server

dev: build
	@./bin/the-button-server

test:
	@go test -v ./...

docker:
	@docker build -t biscuitsbuttonserver .

run: docker
	@docker run -p 3001:3001 biscuitsbuttonserver
