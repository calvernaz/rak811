TEST=go mod download; CGO_ENABLED=0 GOOS=linux go test -v ./...

ARCH ?= amd64

default: test

test:
	@sed \
	    -e 's|COMMAND|$(TEST)|g' \
	    Dockerfile > .dockerfile-$(ARCH)
	@docker build -f .dockerfile-$(ARCH) .
