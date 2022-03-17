all: build

build:
	go build -o ./bin/service ./cmd/service
