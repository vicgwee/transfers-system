install:
	docker pull postgres:14-alpine

postgres:
	mkdir -p postgres-data
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret  -v postgres-data:/var/lib/postgresql/data -d postgres:14-alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root prod

createtables:
	cat db/schema/schema.sql | docker exec -i postgres psql -U root -d prod

dropdb:
	docker exec -it postgres dropdb prod --if-exists

test:
	docker exec -it postgres dropdb testing -f --if-exists
	docker exec -it postgres createdb --username=root --owner=root testing
	cat db/schema/schema.sql | docker exec -i postgres psql -U root -d testing
	go test -v -cover -short ./...

sqlc:
	sqlc generate

mockgen:
	mockgen -package mockdb -destination db/mock/store.go transfers/db/sqlc Store

