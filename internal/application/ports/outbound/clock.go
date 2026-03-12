package outbound

import "time"

type Clock interface {
	NowUTC() time.Time
}
