VERSION := 0.1.7

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o centinela ./cmd/centinela

install:
	go build -ldflags "-X main.Version=$(VERSION)" -o $(HOME)/.local/bin/centinela ./cmd/centinela
