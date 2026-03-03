package system

import "time"

type Clock struct{}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) NowUTC() time.Time {
	return time.Now().UTC()
}
