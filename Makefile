all: build

.PHONY: build
build:
	go build -mod=vendor -o ./bin/web cmd/web/main.go

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
