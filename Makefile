.PHONY: bootstrap generate test test-integration lint build docker-build

bootstrap:
	cd web && pnpm install --frozen-lockfile

generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
	cd web && pnpm generate:api

test:
	go test ./...
	cd web && pnpm test

test-integration:
	test -n "$(ATLAS_TEST_DATABASE_URL)"
	go test -p 1 -tags=integration ./...

lint:
	test -z "$$(gofmt -l cmd internal migrations)"
	go vet ./...
	cd web && pnpm lint && pnpm typecheck

build:
	cd web && pnpm build
	go build ./cmd/atlas

docker-build:
	docker build --build-arg VITE_API_BASE_URL= -t atlas:local .
