package core

import "time"

type TimeProvider interface {
	Now() time.Time
	NewTicker(d time.Duration) Ticker
}

type SystemTimeProvider struct{}

func NewSystemTimeProvider() SystemTimeProvider {
	return SystemTimeProvider{}
}

func (tp SystemTimeProvider) Now() time.Time {
	return time.Now()
}
func (tp SystemTimeProvider) NewTicker(d time.Duration) Ticker {
	return NewSystemTicker(d)
}

type Ticker interface {
	Stop()
	Reset()
	Chan() <-chan time.Time
}

type SystemTicker struct {
	t *time.Ticker
	d time.Duration
}

func NewSystemTicker(d time.Duration) *SystemTicker {
	return &SystemTicker{
		t: time.NewTicker(d),
	}
}
func (st *SystemTicker) Stop() {
	st.t.Stop()
}

func (st *SystemTicker) Reset() {
	st.t.Reset(st.d)
}

func (st *SystemTicker) Chan() <-chan time.Time {
	return st.t.C
}
