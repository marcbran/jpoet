
default:
    @just --list

build:
	go build -v ./...

install: build
	go install -v ./...
