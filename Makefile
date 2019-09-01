GOFILES = $(shell find . -name '*.go')
GO_ENV=GOOS=`go env GOOS` GOARCH=`go env GOARCH` CGO_ENABLED=0

default: test

test: $(GOFILES)
	${GO_ENV} go test ./...
