migrate:
	@export DATABASE_URL="postgres://periscope_dev:${PERISCOPE_POSTGRES_PASSWORD}@localhost:5432/periscope_dev?sslmode=disable"; atlas migrate apply --env gorm

migrations-list:
	atlas migrate ls --env gorm

migrations-generate-diff:
	atlas migrate diff --env gorm

lint:
	golangci-lint run

test:
	go test -v ./...
