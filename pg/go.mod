module github.com/joaofbantunes/outboxkit-go-poc/pg

go 1.25

require (
	github.com/jackc/pgx/v5 v5.7.5
	github.com/joaofbantunes/outboxkit-go-poc/core v0.0.1
)

replace github.com/joaofbantunes/outboxkit-go-poc/core => ../core
