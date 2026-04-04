
test-jsonnet:
    #!/usr/bin/env bash
    set -eu

    jpoet test ./internal
    jpoet test ./pkg

test-go:
    #!/usr/bin/env bash
    set -eu

    go test -v -cover -timeout=120s -parallel=10 ./...

test: test-jsonnet test-go

lint-go:
    #!/usr/bin/env bash
    set -eu

    golangci-lint run

lint: lint-go

build: test lint
    #!/usr/bin/env bash
    set -eu

    mkdir -p dist
    go build -o ./dist -v ./...

it: build
    #!/usr/bin/env bash
    set -eu

    pushd ./examples/build && just it && popd

install: build
    #!/usr/bin/env bash
    set -eu

    go install -v ./...
