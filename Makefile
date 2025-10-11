.PHONY: all
all: generate format test build lint 

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	go build -o bin/server ./cmd/server

.PHONY: docker-compose-up
docker-compose-up:
	bash scripts/generate-dev-certs.sh
	docker compose up -d --build
	docker run --rm --network admin-api_default \
		-v $(PWD)/sql/schema.sql:/schema.sql \
		arigaio/atlas:latest schema apply \
		--url "postgres://admin_api:password@postgresql:5432/tacokumo_admin_db?sslmode=disable" \
		--dev-url "postgres://postgres:password@postgresql:5432/postgres?sslmode=disable" \
		--to "file:///schema.sql" --auto-approve

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
