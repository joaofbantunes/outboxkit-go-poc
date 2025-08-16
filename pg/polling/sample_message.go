package polling

import "time"

// just for PoC, final implementation should allow the user to bring their own message type

type SampleMessage struct {
	ID           int64
	Type         string
	Payload      []byte
	CreatedAt    time.Time
	TraceContext []byte
}
