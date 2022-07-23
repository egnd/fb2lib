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

build: build-index build-summary build-server build-converter ## Build

build-index: ## Build index binary
	@mkdir -p bin/$(GOOS)-$(GOARCH) && rm -f bin/$(GOOS)-$(GOARCH)/build_index
	CGO_ENABLED=0 go build -mod=vendor -ldflags "-X 'main.appVersion=$(BUILD_VERSION)-$(GOOS)-$(GOARCH)'" -o bin/$(GOOS)-$(GOARCH)/build_index cmd/index/*
	@chmod +x bin/$(GOOS)-$(GOARCH)/build_index && ls -lah bin/$(GOOS)-$(GOARCH)/build_index

build-summary: ## Build summary binary
	@mkdir -p bin/$(GOOS)-$(GOARCH) && rm -f bin/$(GOOS)-$(GOARCH)/build_summary
	CGO_ENABLED=0 go build -mod=vendor -ldflags "-X 'main.appVersion=$(BUILD_VERSION)-$(GOOS)-$(GOARCH)'" -o bin/$(GOOS)-$(GOARCH)/build_summary cmd/summary/*
	@chmod +x bin/$(GOOS)-$(GOARCH)/build_summary && ls -lah bin/$(GOOS)-$(GOARCH)/build_summary

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

build-image: ## Build app image
	docker build --tag=fb2lib:debug --build-arg TARGETOS=linux --build-arg TARGETARCH=amd64 .

compose: compose-stop ## Run app
	docker-compose up --build --abort-on-container-exit --renew-anon-volumes

compose-stop: ## Stop app
	docker-compose down --remove-orphans --volumes

_pprof:
	@mkdir -p var/pprof

pprof-for: _pprof
	go tool pprof -svg $(filter-out $@,$(MAKECMDGOALS)) > var/pprof/graph.svg

pprof-save-http: _pprof
	wget http://localhost/debug/pprof/profile -O var/pprof/cpu.prof
	wget http://localhost/debug/pprof/allocs -O var/pprof/allocs.prof
	wget http://localhost/debug/pprof/block -O var/pprof/block.prof
	wget http://localhost/debug/pprof/goroutine -O var/pprof/goroutine.prof
	wget http://localhost/debug/pprof/heap -O var/pprof/heap.prof
	wget http://localhost/debug/pprof/mutex -O var/pprof/mutex.prof	
	wget http://localhost/debug/pprof/threadcreate -O var/pprof/threadcreate.prof

pprof-svg:
	go tool pprof -svg var/pprof/cpu.prof > var/pprof/cpu.svg
	go tool pprof -svg var/pprof/allocs.prof > var/pprof/allocs.svg
	go tool pprof -svg var/pprof/block.prof > var/pprof/block.svg
	go tool pprof -svg var/pprof/goroutine.prof > var/pprof/goroutine.svg
	go tool pprof -svg var/pprof/heap.prof > var/pprof/heap.svg
	go tool pprof -svg var/pprof/mutex.prof > var/pprof/mutex.svg
	go tool pprof -svg var/pprof/threadcreate.prof > var/pprof/threadcreate.svg
