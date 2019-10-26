.PHONY: all
all: vet lint test

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	./scripts/golint.sh ./...

.PHONY: test
test:
	go test ./...
