.PHONY: build build-race clean run run-race test test-race

build:
	go build -o bin/serve cmd/serve/*.go

build-race:
	go build -race -o bin/serve cmd/serve/*.go

clean:
	rm -rf ./bin

run: clean build
	bin/serve

run-race: clean build-race
	bin/serve

test: clean build
	go test -v ./internal/*
	go test -v ./pkg/*
	go run test/test.go

test-race: clean build-race
	go test -race -v ./internal/*
	go test -race -v ./pkg/*
	go run test/test.go
