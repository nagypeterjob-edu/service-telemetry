test:
	go test -count=1 -race -cover -v $(shell go list ./... | grep -v -e /vendor/)

build:
	GOOS=darwin go build -o "bin/service" main.go

build-linux:
	GOOS=linux go build -o "bin/service" main.go

vendor:
	go mod vendor

tidy:
	go mod tidy