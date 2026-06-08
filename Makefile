# Mirante — dev tasks.
# Go is not required on the host: Go targets run inside the official golang
# image via Docker. The web targets use the local Node toolchain.

GO_IMAGE   := golang:1.25-alpine
LINT_IMAGE := golangci/golangci-lint:v2.1.6
API_DIR    := apps/api
WEB_DIR    := apps/web

# Run a command inside the golang image with the module cache cached in a volume.
DOCKER_GO = docker run --rm \
	-v "$(CURDIR)":/src -w /src/$(API_DIR) \
	-v mirante-gocache:/go/pkg/mod \
	-e GOFLAGS=-buildvcs=false \
	$(GO_IMAGE)

.PHONY: help
help: ## List targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## go mod tidy (in container)
	$(DOCKER_GO) go mod tidy

.PHONY: api-build
api-build: ## Build the API
	$(DOCKER_GO) go build ./...

.PHONY: api-vet
api-vet: ## go vet
	$(DOCKER_GO) go vet ./...

.PHONY: api-test
api-test: ## Run Go tests (skips testcontainers paths by default)
	$(DOCKER_GO) go test ./...

.PHONY: api-lint
api-lint: ## Run golangci-lint
	docker run --rm -v "$(CURDIR)":/src -w /src/$(API_DIR) \
		-v mirante-gocache:/go/pkg/mod $(LINT_IMAGE) golangci-lint run

.PHONY: web-install
web-install: ## Install web deps
	cd $(WEB_DIR) && npm install

.PHONY: web-build
web-build: ## Build the web app
	cd $(WEB_DIR) && npm run build

.PHONY: web-dev
web-dev: ## Run the web dev server
	cd $(WEB_DIR) && npm run dev

.PHONY: dev
dev: ## Bring up the full stack (sqld + api + web)
	docker compose up --build

.PHONY: down
down: ## Tear down the stack
	docker compose down
