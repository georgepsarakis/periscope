migrate:
	DATABASE_URL="postgres://periscope_dev:${PERISCOPE_POSTGRES_PASSWORD}@localhost:5432/periscope_dev?sslmode=disable" atlas migrate apply --env gorm

migrations-list:
	atlas migrate ls --env gorm
