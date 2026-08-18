.PHONY: build test run docker
build:
	go build ./cmd/nzbharbor
test:
	go test ./...
run:
	NZBHARBOR_CONFIG=$$(pwd)/config NZBHARBOR_DOWNLOADS=$$(pwd)/downloads go run ./cmd/nzbharbor
docker:
	docker compose up -d --build
