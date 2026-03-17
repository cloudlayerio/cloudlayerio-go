.PHONY: test test-coverage lint vet fmt check smoke

test:
	go test -race -count=1 ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

check: vet lint test

smoke:
	go test -tags=smoke -race -count=1 -v ./...
