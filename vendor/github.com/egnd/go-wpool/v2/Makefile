#!make

MAKEFLAGS += --always-make

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
	@clear && rm -f interfaces/mock_*.go
	mockery

tests: ## Run unit tests
	@rm -rf coverage && mkdir -p coverage
	CGO_ENABLED=1 go test -mod=vendor -race -cover -covermode=atomic -coverprofile=coverage/profile.out ./...

benchmarks: ## Run benchmarks
	@clear
	go test -mod=vendor -benchmem -bench . benchmarks_test.go

coverage: tests ## Check code coveragem
	cat coverage/profile.out | grep -v "mock_" > coverage/profile.cov
	go tool cover -func=coverage/profile.cov
	go tool cover -html=coverage/profile.cov -o coverage/report.html

lint: ## Lint source code
	@clear
	golangci-lint run --color=always --config=.golangci.yml ./...

########################################################################################################################

docker-lint:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint make golangci/golangci-lint:v1.45 lint

docker-mocks:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint sh vektra/mockery:v2 -c "apk add -q make && make mocks"

docker-tests:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint make golang:1.18 tests

docker-coverage:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint make golang:1.18 coverage
	@echo "Read report at file://$$(pwd)/coverage/report.html"

docker-benchmarks:
	docker run --rm -it -v $$(pwd):/src -w /src --entrypoint make golang:1.18 benchmarks
