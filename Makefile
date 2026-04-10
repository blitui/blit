.PHONY: test lint build cover tidy clean fmt vet snapshot

test:
	go test -race ./...

lint:
	golangci-lint run ./...

build:
	go build ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

tidy:
	go mod tidy

snapshot:
	goreleaser build --snapshot --clean

clean:
	rm -f coverage.out coverage.html
	rm -f *.exe
