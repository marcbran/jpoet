tool := "go tool -modfile=tools/go.mod"

test-go:
    go test -v -cover -timeout=120s -parallel=10 ./...

test-race:
    go test -v -race -timeout=120s ./...

test-jsonnet:
    go run . test ./internal
    go run . test ./pkg

test: test-go test-jsonnet

test-e2e:
    go test -v -race -tags e2e -timeout=300s ./tests/...

lint:
    {{tool}} golangci-lint run ./...

vuln:
    {{tool}} govulncheck ./...

build:
    mkdir -p dist
    go build -o ./dist -v ./...

it: build
    cd examples/build && just it

install:
    go install -v ./...

ci: lint vuln test test-race build test-e2e
