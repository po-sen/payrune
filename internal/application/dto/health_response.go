package dto

import "time"

type HealthResponse struct {
	Status    string
	Timestamp time.Time
}
