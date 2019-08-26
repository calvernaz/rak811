GOFILES = $(shell find . -name '*.go')

default: test

test: $(GOFILES)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test ./...
