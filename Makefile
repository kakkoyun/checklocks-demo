# Makefile for checklocks-demo

# Go paths
GOPATH=$(shell go env GOPATH)
VETTOOL=$(GOPATH)/bin/checklocks

# Phony targets
.PHONY: all install-vettool lint test clean

# Default target
all: lint test

# Install the checklocks vet tool
install-vettool:
	@echo "Installing checklocks vet tool..."
	@go install gvisor.dev/gvisor/tools/checklocks/cmd/checklocks@latest
	@echo "Vet tool installed to $(VETTOOL)"

# Run go vet with the checklocks analyzer
# Ensure the vet tool is installed first.
lint: install-vettool
	@echo "Running checklocks linter with debug tag..."
	@go vet -vettool=$(VETTOOL) -tags debug ./...

# Run tests
test:
	@echo "Running tests with race detector, timeout, and debug tag..."
	@go test -race -timeout 30s -tags debug ./...

# Clean build artifacts (optional)
clean:
	@echo "Cleaning..."
	@go clean
