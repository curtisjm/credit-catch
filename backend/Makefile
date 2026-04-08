.PHONY: build run test test-integration lint migrate-up migrate-down db-up db-down clean

BINARY   := creditcatch-server
DB_URL   ?= $(DATABASE_URL)

build:
	go build -o bin/$(BINARY) ./cmd/server

run: build
	DATABASE_URL=$(DB_URL) JWT_SECRET=dev-secret-change-me ./bin/$(BINARY)

test:
	go test ./... -v -race -count=1

test-integration:
	DATABASE_URL=$(DB_URL) JWT_SECRET=test-secret go test ./... -tags=integration -v -race -count=1

lint:
	go vet ./...

migrate-up:
	@for f in migrations/*.up.sql; do \
		echo "Applying $$f..."; \
		psql "$(DB_URL)" -f "$$f"; \
	done

migrate-down:
	@for f in $$(ls -r migrations/*.down.sql); do \
		echo "Reverting $$f..."; \
		psql "$(DB_URL)" -f "$$f"; \
	done

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

clean:
	rm -rf bin/
