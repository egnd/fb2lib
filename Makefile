#!make

MAKEFLAGS += --always-make
BUILD_VERSION=dev
FB2C_VERSION=v1.60.2.2

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

%:
	@:

########################################################################################################################

owner: ## Reset folder owner
	sudo chown --changes -R $$(whoami) ./
	@echo "Success"

build: build-indexer build-server build-converter ## Build

build-indexer: ## Build indexer
	@mkdir -p bin/$(GOOS)-$(GOARCH) && rm -f bin/$(GOOS)-$(GOARCH)/indexer
	CGO_ENABLED=0 go build -mod=vendor -ldflags "-X 'main.appVersion=$(BUILD_VERSION)-$(GOOS)-$(GOARCH)'" -o bin/$(GOOS)-$(GOARCH)/indexer cmd/indexer/*
	@chmod +x bin/$(GOOS)-$(GOARCH)/indexer && ls -lah bin/$(GOOS)-$(GOARCH)/indexer

build-server: ## Build server
	@mkdir -p bin/$(GOOS)-$(GOARCH) && rm -f bin/$(GOOS)-$(GOARCH)/server
	CGO_ENABLED=0 go build -mod=vendor -ldflags "-X 'main.appVersion=$(BUILD_VERSION)-$(GOOS)-$(GOARCH)'" -o bin/$(GOOS)-$(GOARCH)/server cmd/server/*
	@chmod +x bin/$(GOOS)-$(GOARCH)/server && ls -lah bin/$(GOOS)-$(GOARCH)/server

build-converter: ## Build server
	@mkdir -p bin/$(GOOS)-$(GOARCH) && rm -f bin/$(GOOS)-$(GOARCH)/fb2c
ifeq ($(wildcard fb2c-$(GOOS)-$(GOARCH)-$(FB2C_VERSION).zip),)
	wget https://github.com/egnd/fb2converter/releases/download/$(FB2C_VERSION)/fb2c-$(GOOS)-$(GOARCH)-$(FB2C_VERSION).zip
endif
	unzip fb2c-$(GOOS)-$(GOARCH)-$(FB2C_VERSION).zip
	mv fb2c bin/$(GOOS)-$(GOARCH)/fb2c && ls -lah bin/$(GOOS)-$(GOARCH)/fb2c

compose: compose-stop ## Run app
	docker-compose up --build --abort-on-container-exit --renew-anon-volumes

compose-stop: ## Stop app
ifeq ($(wildcard .env),)
	cp .env.dist .env
endif
	docker-compose down --remove-orphans --volumes

compose-index:
	docker-compose exec server bin/indexer

_pprof:
	@mkdir -p var/pprof

pprof-for: _pprof
	go tool pprof -svg $(filter-out $@,$(MAKECMDGOALS)) > var/pprof/graph.svg

pprof-http: _pprof
	go tool pprof -svg http://localhost:8080/debug/pprof/profile > var/pprof/cpu.svg
	go tool pprof -svg http://localhost:8080/debug/pprof/allocs > var/pprof/allocs.svg
	go tool pprof -svg http://localhost:8080/debug/pprof/block > var/pprof/block.svg
	go tool pprof -svg http://localhost:8080/debug/pprof/goroutine > var/pprof/goroutine.svg
	go tool pprof -svg http://localhost:8080/debug/pprof/heap > var/pprof/heap.svg
	go tool pprof -svg http://localhost:8080/debug/pprof/mutex > var/pprof/mutex.svg	
	go tool pprof -svg http://localhost:8080/debug/pprof/threadcreate > var/pprof/threadcreate.svg
