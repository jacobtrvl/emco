all: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -tags netgo -o sacontroller ./cmd/sacontroller/main.go

build: clean test cover
deploy: build

.PHONY: test
test: clean
	@go test -race ./...

format:
	@go fmt ./...

clean:
	@find . -name "*so" -delete
	 @rm -f ../../bin/orchestrator/orchestrator coverage.html coverage.out

.PHONY: cover
cover:
	@go test -race ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html