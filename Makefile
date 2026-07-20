# No search_path: golang-migrate connects before schema voco exists (see 000001).
MIGRATE_PG_URL ?= postgres://voco:voco@localhost:5432/voco?sslmode=disable

.PHONY: migrate-up migrate-down migrate-version migrate-steps
migrate-up:
	VOCO_PG_URL="$(MIGRATE_PG_URL)" VOCO_PG_MIGRATIONS_PATH=migrations VOCO_RELEASE_MODE=dev go run ./cmd/migrate up

migrate-down:
	VOCO_PG_URL="$(MIGRATE_PG_URL)" VOCO_PG_MIGRATIONS_PATH=migrations VOCO_RELEASE_MODE=dev go run ./cmd/migrate down

migrate-version:
	VOCO_PG_URL="$(MIGRATE_PG_URL)" VOCO_PG_MIGRATIONS_PATH=migrations VOCO_RELEASE_MODE=dev go run ./cmd/migrate version

migrate-steps:
	@test -n "$(N)" || (echo "usage: make migrate-steps N=-1"; exit 1)
	VOCO_PG_URL="$(MIGRATE_PG_URL)" VOCO_PG_MIGRATIONS_PATH=migrations VOCO_RELEASE_MODE=dev go run ./cmd/migrate steps $(N)
