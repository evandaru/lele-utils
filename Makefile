BINARY := lele-dev
CMD := ./cmd/lele-dev
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
	rm -f $(BINARY) lele-dev-linux-amd64 lele-dev-linux-arm64 lele-dev-darwin-amd64 lele-dev-darwin-arm64

# PRD §31 — release binaries
cross:
	GOOS=linux GOARCH=amd64 $(GO) build -o lele-dev-linux-amd64 $(CMD)
	GOOS=linux GOARCH=arm64 $(GO) build -o lele-dev-linux-arm64 $(CMD)
	GOOS=darwin GOARCH=amd64 $(GO) build -o lele-dev-darwin-amd64 $(CMD)
	GOOS=darwin GOARCH=arm64 $(GO) build -o lele-dev-darwin-arm64 $(CMD)
