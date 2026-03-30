package pipeline

import "time"

type Meta struct {
	Created time.Time
	TTL     time.Duration
}

func (m Meta) Expired(now time.Time) bool {
	return !m.Created.Add(m.TTL).After(now)
}
