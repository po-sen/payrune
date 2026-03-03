package out

import "time"

type Clock interface {
	NowUTC() time.Time
}
