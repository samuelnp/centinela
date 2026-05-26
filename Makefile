VERSION := 0.3.6

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o centinela ./cmd/centinela

install:
	go build -ldflags "-X main.Version=$(VERSION)" -o $(HOME)/.local/bin/centinela ./cmd/centinela

test-cover:
	./scripts/check-coverage.sh
