all: build

build:
	go build -o ./bin/iino-service ./cmd/service
