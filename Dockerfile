FROM golang:1.21.0-bookworm
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
EXPOSE 3001
RUN CGO_ENABLED=1 go build -o bin/the-button-server
CMD ["./bin/the-button-server"]
