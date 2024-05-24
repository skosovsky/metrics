#.SILENT:
APP=metrics

.PHONY: help
help: Makefile ## Show this help
	@echo
	@echo "Choose a command run in "$(APP)":"
	@echo
	@fgrep -h "##" $(MAKEFILE_LIST) | sed -e 's/\(\:.*\#\#\)/\:\ /' | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: build-apps
build-apps: ## Build an application
	@echo "Building ${APP} ..."
	mkdir -p build
	go build -o build/server metrics/cmd/server
	go build -o build/agent metrics/cmd/agent
	go generate ./...

.PHONY: build-test
build-test: ## Build an application
	@echo "Building ${APP} ..."
	cd cmd/server && go build -buildvcs=false -o server
	cd cmd/agent && go build -buildvcs=false  -o agent
	go generate ./...

test-static: ## Test static
	@echo "Testing ${APP} - static..."
	go vet -vettool="$(shell which ./tests/statictest-darwin-arm64)" ./...

test1: ## Test increment #1
	@echo "Testing ${APP} - increment 1..."
	tests/metricstest-darwin-arm64 -test.v -test.run="^TestIteration1$$" -binary-path=cmd/server/server

test2: ## Test increment #2
	@echo "Testing ${APP} - increment 2..."
	tests/metricstest-darwin-arm64 -test.v -test.run="^TestIteration2[AB]*$$" -source-path=. -agent-binary-path=cmd/agent/agent

test3: ## Test increment #3
	@echo "Testing ${APP} - increment 3..."
	tests/metricstest-darwin-arm64 -test.v -test.run="^TestIteration3[AB]*$$" -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server

test4: ## Test increment #4
	@echo "Testing ${APP} - increment 4..."
	tests/metricstest-darwin-arm64 -test.v -test.run="^TestIteration4$$" -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8001 -source-path=.

run: ## Run an application
	@echo "Starting ${APP} ..."
	go run main.go

test: ## Run an application
	@echo "Testing ${APP} ..."
	go test

bench: ## Run an application
	@echo "Benchmarking ${APP} ..."
	go test -bench=. .

clean: ## Clean a garbage
	@echo "Cleaning"
	go clean
	rm -rf build

lint: ## Check a code by golangci-lint
	@echo "Linter checking..."
	golangci-lint run -c .golangci.yml ./...
