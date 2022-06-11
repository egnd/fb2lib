#!make

MAKEFLAGS += --always-make
CALL_PARAM=$(filter-out $@,$(MAKECMDGOALS))

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

%:
	@:

########################################################################################################################

owner: ## Reset folder owner
	sudo chown --changes -R $$(whoami) ./
	@echo "Success"

check-conflicts: ## Find git conflicts
	@if grep -rn '^<<<\<<<< ' .; then exit 1; fi
	@if grep -rn '^===\====$$' .; then exit 1; fi
	@if grep -rn '^>>>\>>>> ' .; then exit 1; fi
	@echo "All is OK"

check-todos: ## Find TODO's
	@if grep -rn '@TO\DO:' .; then exit 1; fi
	@echo "All is OK"

check-master: ## Check for latest master in current branch
	@git remote update
	@if ! git log --pretty=format:'%H' | grep $$(git log --pretty=format:'%H' -n 1 origin/master) > /dev/null; then exit 1; fi
	@echo "All is OK"

mocks: ## Generate mocks
	@rm -rf mocks
	mockery --name=.

tests: ## Run unit tests
	@mkdir -p profiles
	CGO_ENABLED=1 go test -mod=vendor -race -cover -covermode=atomic -coverprofile=profiles/cover.out.tmp ./...

benchmarks: ## Run benchmarks
	go test -mod=vendor -benchmem -bench . ./...

coverage: tests ## Check code coveragem
	@cat profiles/cover.out.tmp | grep -v "mock_" > profiles/cover.out
	go tool cover -func=profiles/cover.out
	go tool cover -html=profiles/cover.out -o profiles/report.html

lint: ## Lint source code
	golangci-lint run --color=always --config=.golangci.yml ./...

profiles:
	@mkdir -p profiles
	go test -cpuprofile profiles/cpu.out -memprofile profiles/mem.out -bench . $(CALL_PARAM)/benchmarks_test.go
	go tool pprof -svg go-pipeline.test profiles/cpu.out > profiles/cpu.svg
	go tool pprof -svg -alloc_space go-pipeline.test profiles/mem.out > profiles/mem.svg
	go tool pprof -svg -alloc_objects go-pipeline.test profiles/mem.out > profiles/obj.svg

########################################################################################################################

docker-lint:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint make golangci/golangci-lint:v1.45 lint

docker-tests:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint make golang:1.18 tests

docker-coverage:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint make golang:1.18 coverage
	@echo "Read report at file://$$(pwd)/profiles/report.html"

docker-profiles:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint bash golang:1.18 -c "apt-get update -qq && apt-get install -y graphviz > /dev/null && make profiles $(CALL_PARAM)"
	@echo "Read reports at:"
	@echo "- file://$$(pwd)/profiles/cpu.svg"
	@echo "- file://$$(pwd)/profiles/mem.svg"
	@echo "- file://$$(pwd)/profiles/obj.svg"

docker-benchmarks:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint make golang:1.18 benchmarks

docker-mocks:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint sh vektra/mockery:v2 -c "apk add -q make && make mocks"