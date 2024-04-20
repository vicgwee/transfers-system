## Installation:

Install Go (version >= 1.18) and Docker Desktop
https://go.dev/
https://www.docker.com/products/docker-desktop/

## Setup:
Install Go dependencies:
```
go mod tidy
```

Initialize and run postgres Docker container:
```
make postgres
```
Create DB and Tables
```
make createdb & make createtables
```

## Run:

Run unit tests
```
make test
```

Run server:
```
go run main.go
```

## Sample Curl Requests:
Create two accounts:
```
curl --location 'localhost:8080/accounts' \
--header 'Content-Type: application/json' \
--data '{
    "account_id": 1,
    "initial_balance": "1.23"
}'
curl --location 'localhost:8080/accounts' \
--header 'Content-Type: application/json' \
--data '{
    "account_id": 2,
    "initial_balance": "4.56"
}'
```

Create Transaction:
```
curl --location 'localhost:8080/transactions' \
--header 'Content-Type: application/json' \
--data '{
    "source_account_id":2,
    "destination_account_id": 1,
    "amount": "1"
}'
```

Get Account:
```
curl --location 'localhost:8080/accounts/1'
```


# Assumptions:
All accounts created are cash accounts, balance must be >= 0 (enforced by DB constraint)
AccountID must be >0 (enforced by binding validation check)
Tested with up to 200 concurrent transactions between two accounts (store_test.go)

As an internal transfers system,the server is secure, there is no need to:
    - have a strong database username and password, and encrypt it while it's stored
    - implement authentication or authorization checks
    - encrypt the user data (persisted in postgres-data directory)
The database is reliable, periodic database snapshots and backups are not implemented
During server and database maintenance/upgrades, downtime is acceptable