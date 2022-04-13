#!make

MAKEFLAGS += --always-make
BUILD_VERSION=dev
FB2C_VERSION=v1.60.2

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

%:
	@:

########################################################################################################################

owner: ## Reset folder owner
	sudo chown --changes -R $$(whoami) ./
	@echo "Success"

build-indexer: ## Build indexer
	@mkdir -p bin && rm -f bin/indexer
	CGO_ENABLED=0 go build -mod=vendor -ldflags "-X 'main.appVersion=$(BUILD_VERSION)-$(GOOS)-$(GOARCH)'" -o bin/indexer cmd/indexer/*
	@chmod +x bin/indexer && ls -lah bin/indexer

build-server: ## Build server
	@mkdir -p bin && rm -f bin/server
	CGO_ENABLED=0 go build -mod=vendor -ldflags "-X 'main.appVersion=$(BUILD_VERSION)-$(GOOS)-$(GOARCH)'" -o bin/server cmd/server/*
	@chmod +x bin/server && ls -lah bin/server

build-converter: ## Build server
ifeq ($(wildcard fb2c),)
	git clone https://github.com/rupor-github/fb2converter.git fb2c && cd fb2c && git remote update && git checkout $(FB2C_VERSION)
	mkdir -p fb2c/misc && cp fb2c/cmake/version.go.in fb2c/misc/version.go && \
		sed -i 's/@PRJ_VERSION_MAJOR@.@PRJ_VERSION_MINOR@.@PRJ_VERSION_PATCH@/$(FB2C_VERSION)-$(GOOS)-$(GOARCH)/g' fb2c/misc/version.go && \
		sed -i 's/@GIT_HASH@/SOME_HASH/g' fb2c/misc/version.go
	cd fb2c/static/dictionaries && \
		wget -r -l1 --no-parent -nd -A.pat.txt http://ctan.math.utah.edu/ctan/tex-archive/language/hyph-utf8/tex/generic/hyph-utf8/patterns/txt && \
		wget -r -l1 --no-parent -nd -A.hyp.txt http://ctan.math.utah.edu/ctan/tex-archive/language/hyph-utf8/tex/generic/hyph-utf8/patterns/txt && \
		for a in $$(ls *.txt); do gzip $$a; done
	cd fb2c/static/sentences && \
		curl -L https://api.github.com/repos/neurosnap/sentences/tarball | tar xz --wildcards '*/data/*.json' --strip-components=2 && \
		for a in $$(ls *.json); do gzip $$a; done
endif
	@mkdir -p bin && rm -f bin/fb2c
	cd fb2c && CGO_ENABLED=0 go build -mod=vendor -o ../bin/fb2c fb2c.go
	@chmod +x bin/fb2c && ls -lah bin/fb2c

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
