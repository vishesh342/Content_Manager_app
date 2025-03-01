postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

createdb:
	docker exec -it postgres createdb --username=root --owner=root contentdb

dropdb:
	docker exec -it postgres dropdb contentdb

migrateup:
	migrate -path db/migrations -database "postgres://root:secret@localhost:5432/contentdb?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migrations -database "postgres://root:secret@localhost:5432/contentdb?sslmode=disable" -verbose dowm

sqlc:
	sqlc generate

server:
	go run main.go

.PHONY:postgres createdb dropdb migrateup migratedown sqlc server