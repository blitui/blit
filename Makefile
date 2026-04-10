.PHONY: test lint build cover tidy clean

test:
	go test -race ./...

lint:
	golangci-lint run ./...

build:
	go build ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

tidy:
	go mod tidy

clean:
	rm -f coverage.out coverage.html
	rm -f *.exe
