#!make

MAKEFLAGS += --always-make
BUILD_VERSION=dev

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

%:
	@:

########################################################################################################################

owner: ## Reset folder owner
	sudo chown --changes -R $$(whoami) ./
	@echo "Success"

build: ## Build app
	@rm -rf bin && mkdir bin
	CGO_ENABLED=0 go build -mod=vendor -ldflags "-X 'main.appVersion=$(BUILD_VERSION)-$(GOOS)-$(GOARCH)'" -o bin/server cmd/server/*
	CGO_ENABLED=0 go build -mod=vendor -ldflags "-X 'main.appVersion=$(BUILD_VERSION)-$(GOOS)-$(GOARCH)'" -o bin/indexer cmd/indexer/*
	cp fb2c/$(GOOS)-$(GOARCH) bin/fb2c
	@chmod +x bin/server bin/indexer bin/fb2c && ls -lah bin/server bin/indexer bin/fb2c

compose: compose-stop ## Run app
ifeq ($(wildcard docker-compose.override.yml),)
	ln -s docker-compose.build.yml docker-compose.override.yml
endif
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
