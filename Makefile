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
	go mod tidy
	go generate ./...
	go build -o build/server metrics/cmd/server
	go build -o build/agent metrics/cmd/agent

test-static: ## Test static
	@echo "Testing ${APP} - static..."
	go vet -vettool=$(which ./tests/statictest-darwin-arm64) ./...

.PHONY: test_all lint tests build-test test1 test2 test3 test4 test5 test6 test7 test8
build-test: ## Build an application
	@echo "Building ${APP} ..."
	go mod tidy
	go generate ./...
	cd cmd/server && go build -buildvcs=false -o server
	cd cmd/agent  && go build -buildvcs=false -o agent

lint: ## Check a code by golangci-lint
	@echo "Linter checking..."
	golangci-lint run --fix -c .golangci.yml ./...

tests: ## Internal tests
	@echo "Testing ${APP} ..."
	go test ./...

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
	tests/metricstest-darwin-arm64 -test.v -test.run="^TestIteration4$$" -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8004 -source-path=.

test5: ## Test increment #5
	@echo "Testing ${APP} - increment 5..."
	tests/metricstest-darwin-arm64 -test.v -test.run="^TestIteration5$$" -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8005 -source-path=.

test6: ## Test increment #6
	@echo "Testing ${APP} - increment 6..."
	tests/metricstest-darwin-arm64 -test.v -test.run="^TestIteration6$$" -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8006 -source-path=.

test7: ## Test increment #7
	@echo "Testing ${APP} - increment 7..."
	tests/metricstest-darwin-arm64 -test.v -test.run="^TestIteration7$$" -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8007 -source-path=.

test8: ## Test increment #8
	@echo "Testing ${APP} - increment 8..."
	tests/metricstest-darwin-arm64 -test.v -test.run="^TestIteration8$$" -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8008 -source-path=.

test_all: lint tests build-test test1 test2 test3 test4 test5 test6 test7 test8
	@echo "All tests completed."

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
