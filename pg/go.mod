module github.com/joaofbantunes/outboxkit-go-poc/pg

go 1.25

require (
	github.com/jackc/pgx/v5 v5.7.5
	github.com/joaofbantunes/outboxkit-go-poc/core v0.0.1
)

replace github.com/joaofbantunes/outboxkit-go-poc/core => ../core

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)
