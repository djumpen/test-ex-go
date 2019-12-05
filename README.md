# Golang/Postgres Test Exercise

## About

Application illustrates REST API service with concurrent-ready database interactions

## Run

Run service using docker-compose

`$ docker-compose up`

Endpoint: `http://localhost:8088/event`

Msg example:
```
{"state": "win", "amount": "10.15", "transactionId": "some generated identificator"}
```

## Tasks

To be able to scale our app I moved cancellation task to it's own runner. Repeats can be managed either by our app or by CronJob (depends on config)

`$ go run cmd/task/cancellation.go`

## Testing

Integration tests are done using [testcontainers](https://github.com/testcontainers/testcontainers-go)

`$ go test ./... -count=1`



