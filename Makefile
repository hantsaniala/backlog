BINARY=backlog

build:
	go build -o $(BINARY) .

install:
	go install .

release:
	goreleaser release --clean

test:
	go test ./... -v

clean:
	go clean
	rm -f $(BINARY)

.PHONY: build install release test clean
