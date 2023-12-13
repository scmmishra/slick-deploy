NAME=slick
VERSION=0.0.1

.DEFAULT_GOAL := help

.PHONY: build
## build: Compile the packages.
build:
	@go build -v -o bin/$(NAME) cmd/$(NAME)/main.go

.PHONY: install
## install: Install the packages.
install: build
	@mv bin/$(NAME) $(GOPATH)/bin/$(NAME)

.PHONY: clean
## clean: Clean projects and previous builds
clean:
	@rm -rf bin/*

.PHONY: deps
## deps: Download modules
deps:
	@go mod download

.PHONY: watch
## watch: Reload the app whenever the source changes
watch:
	@which reflex > /dev/null || (go install github.com/cespare/reflex@latest)
	reflex -s -r '\.go$$' make run

.PHONY: test
## test: Run all tests
test:
	@go test -v ./...

.PHONY: help
all: help
## help: show this help message
help: Makefile
	@echo
	@echo " Choose a command to run:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
	@echo
