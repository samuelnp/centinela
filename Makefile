VERSION := 0.1.0

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o centinela ./cmd/centinela

install:
	go install -ldflags "-X main.Version=$(VERSION)" ./cmd/centinela
