package core

type Key struct {
	Provider string
	Client   string
}

func (k Key) String() string {
	return k.Provider + ":" + k.Client
}

func NewKey(provider, client string) Key {
	return Key{
		Provider: provider,
		Client:   client,
	}
}

func NewDefaultKey(provider string) Key {
	return Key{
		Provider: provider,
		Client:   "default",
	}
}
