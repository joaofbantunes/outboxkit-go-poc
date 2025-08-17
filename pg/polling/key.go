package polling

import "github.com/joaofbantunes/outboxkit-go-poc/core"

const provider = "pg_polling"

func NewKey(client string) *core.Key {
	return core.NewKey(provider, client)
}

func NewDefaultKey() *core.Key {
	return core.NewDefaultKey(provider)
}
