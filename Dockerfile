FROM golang:1.21-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
EXPOSE 3001
RUN go build -o bin/the-button-server
CMD ["./bin/the-button-server"]
