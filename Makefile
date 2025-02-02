#-include .env

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
all, help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nMakefile help:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

generate: generate_grpc ### generate all
	echo "generate..."
.PHONY: generate

generate_grpc: ### generate grpc
	cd ./proto && \
	buf dep update && \
	buf generate
.PHONY: generate_grpc

sec: ### run gosec
	gosec -exclude=G103,G115,G404,G402 ./...
.PHONY: sec

deps: deps_go ### install dependencies
	echo "deps..."
.PHONY: deps

deps_go: ### install dependencies from go
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@-GOBIN=$(GOBIN) go install github.com/bufbuild/buf/cmd/buf@latest
.PHONY: deps_brew

test: ### run tests
	go test -v ./...
.PHONY: test