.PHONY: test lint build cover tidy clean fmt fix vet snapshot check

test:
	go test -race ./...

lint:
	golangci-lint run ./...

build:
	go build ./...

fmt:
	@gofmt -l . | grep -q . && echo "files need gofmt" && exit 1 || echo "fmt ok"

fix:
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

check:
	@$(MAKE) fmt vet lint test

clean:
	rm -f coverage.out coverage.html cover.out cover2.out coverage_new.out charts_cover.out core.out *.exe
