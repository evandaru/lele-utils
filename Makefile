BINARY := dev-utils
CMD := ./cmd/dev-utils
GO := go

.PHONY: run build install test vet fmt clean cross version

run:
	$(GO) run $(CMD)

build:
	$(GO) build -o $(BINARY) $(CMD)

install:
	$(GO) install $(CMD)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -l -w .

clean:
	rm -f $(BINARY) dev-utils-linux-amd64 dev-utils-linux-arm64 dev-utils-darwin-amd64 dev-utils-darwin-arm64

# PRD §31 — release binaries
cross:
	GOOS=linux GOARCH=amd64 $(GO) build -o dev-utils-linux-amd64 $(CMD)
	GOOS=linux GOARCH=arm64 $(GO) build -o dev-utils-linux-arm64 $(CMD)
	GOOS=darwin GOARCH=amd64 $(GO) build -o dev-utils-darwin-amd64 $(CMD)
	GOOS=darwin GOARCH=arm64 $(GO) build -o dev-utils-darwin-arm64 $(CMD)
