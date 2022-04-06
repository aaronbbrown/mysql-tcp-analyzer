GO=go
FILES := $(shell go list ./.../)

.PHONY: all test dep clean

all: test
	mkdir -p bin
	$(GO) build -o bin/analyze

test:
	$(GO) test -race -v $(FILES) -cover -coverprofile=coverage.out

clean:
	rm -vrf bin

dep:
	$(GO) get
	$(GO) mod tidy
