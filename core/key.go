package core

type Key struct {
	provider string
	client   string
	s        string
}

func (k *Key) Provider() string {
	return k.provider
}

func (k *Key) Client() string {
	return k.client
}

func (k *Key) String() string {
	return k.s
}

func NewKey(provider, client string) *Key {
	return &Key{
		provider: provider,
		client:   client,
		s:        provider + ":" + client,
	}
}

func NewDefaultKey(provider string) *Key {
	return &Key{
		provider: provider,
		client:   "default",
		s:        provider + ":default",
	}
}
