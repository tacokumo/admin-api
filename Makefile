.PHONY: all
all: generate format test build lint

.PHONY: format
format:
	go fmt ./...

# Test commands based on ADR-001 test strategy
.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

# Migration configuration
IS_DOCKER ?= true
DB_HOST ?= postgresql
DB_PORT ?= 5432
DB_USER ?= admin_api
DB_PASSWORD ?= password
DB_NAME ?= tacokumo_admin_db
DEV_DB_USER ?= postgres
DEV_DB_PASSWORD ?= password
DEV_DB_NAME ?= postgres
DOCKER_NETWORK ?= admin-api_default

# Conditional network flag for Docker
ifeq ($(IS_DOCKER),true)
	NETWORK_FLAG = --network $(DOCKER_NETWORK)
else
	NETWORK_FLAG = --network host
endif

.PHONY: migrate
migrate:
	docker run --rm $(NETWORK_FLAG) \
		-v $(PWD)/sql/schema.sql:/schema.sql \
		arigaio/atlas:latest schema apply \
		--url "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" \
		--dev-url "postgres://$(DEV_DB_USER):$(DEV_DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DEV_DB_NAME)?sslmode=disable" \
		--to "file:///schema.sql" --auto-approve

.PHONY: docker-compose-up
docker-compose-up:
	bash scripts/generate-dev-certs.sh
	docker compose up -d --build
	$(MAKE) migrate

.PHONY: docker-compose-down
docker-compose-down:
	docker compose down --volumes

.PHONY: generate
generate:
	go tool ogen -clean -package generated -target ./pkg/apis/v1alpha1/generated ./api-spec/admin/v1alpha1/openapi.yaml
	go tool sqlc generate -f ./sql/sqlc.yaml

.PHONY: lint
lint:
	@which golangci-lint > /dev/null || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b tools v2.5.0)
	./tools/golangci-lint run

.PHONY: submodule
submodule:
	git submodule update --init --recursive
	git submodule update --remote

# Development environment management
.PHONY: verify-setup
verify-setup:
	bash scripts/verify-setup.sh

.PHONY: reset-dev-env
reset-dev-env:
	bash scripts/reset-dev-env.sh

.PHONY: setup-env
setup-env:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env from template. Please edit it with your GitHub OAuth credentials."; \
	else \
		echo ".env already exists"; \
	fi
