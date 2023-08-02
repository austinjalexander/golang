GOBIN := $(shell pwd)/bin

.PHONY: all install mod run

all:
	go build -o $(GOBIN)/onetab main.go

mod:
	go mod tidy
	go mod vendor

run:
	go run main.go

run-build:
	$(GOBIN)/onetab