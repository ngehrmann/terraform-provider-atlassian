default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Build the provider
.PHONY: build
build:
	go build -o terraform-provider-atlassian

# Install the provider locally for development
.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/atlassian/0.1.0/linux_amd64
	cp terraform-provider-atlassian ~/.terraform.d/plugins/registry.terraform.io/hashicorp/atlassian/0.1.0/linux_amd64/

# Generate docs
.PHONY: docs
docs:
	go generate ./...

# Format code
.PHONY: fmt
fmt:
	gofmt -s -w .
	terraform fmt -recursive ./examples/

# Lint
.PHONY: lint
lint:
	golangci-lint run

# Test
.PHONY: test
test:
	go test -v ./...

# Clean
.PHONY: clean
clean:
	rm -f terraform-provider-atlassian

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build     - Build the provider"
	@echo "  install   - Install the provider locally"
	@echo "  test      - Run unit tests"
	@echo "  testacc   - Run acceptance tests"
	@echo "  docs      - Generate documentation"
	@echo "  fmt       - Format code"
	@echo "  lint      - Run linter"
	@echo "  clean     - Clean build artifacts"
	@echo "  help      - Show this help message"