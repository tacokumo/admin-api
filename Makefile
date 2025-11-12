.PHONY: all
all: generate format test build lint

.PHONY: test-all
test-all: test scenario-test

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: scenario-test
scenario-test:
	go tool ginkgo -vv ./test/scenario

.PHONY: build
build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

# Migration configuration
IS_DOCKER ?= true
HOST ?= postgresql
PORT ?= 5432
USER ?= admin_api
PASSWORD ?= password
DB ?= tacokumo_admin_db
DEV_USER ?= postgres
DEV_PASSWORD ?= password
DEV_DB ?= postgres
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
		--url "postgres://$(USER):$(PASSWORD)@$(HOST):$(PORT)/$(DB)?sslmode=disable" \
		--dev-url "postgres://$(DEV_USER):$(DEV_PASSWORD)@$(HOST):$(PORT)/$(DEV_DB)?sslmode=disable" \
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
