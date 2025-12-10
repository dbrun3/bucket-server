.PHONY: build clean run release

ifneq (,$(wildcard .env))
  include .env
  export
endif

# Default value if not set in .env or shell
SERVER_URL ?= localhost:8080

build:
	go build -o bin/bs ./main.go

clean:
	rm -rf bin/*

run:
	go run main.go

OLD_VERSION := $(shell git describe --tags --abbrev=0)
NEW_VERSION := $(shell svu next)

release:
	git tag $(NEW_VERSION); \
	git push origin main --tags