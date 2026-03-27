package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	inport "payrune/internal/application/ports/inbound"
)

type fakeClock struct {
	now time.Time
}

func (f *fakeClock) NowUTC() time.Time {
	return f.now
}

func TestCheckHealthUseCaseExecute(t *testing.T) {
	expected := time.Date(2026, time.March, 3, 11, 30, 0, 0, time.UTC)
	useCase := NewCheckHealthUseCase(&fakeClock{now: expected})

	response, err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if response.Status != "up" {
		t.Fatalf("unexpected status: got %s", response.Status)
	}

	if response.Timestamp != expected.Format(time.RFC3339) {
		t.Fatalf("unexpected timestamp: got %s", response.Timestamp)
	}
}

func TestCheckHealthUseCaseValidationMissingClock(t *testing.T) {
	useCase := NewCheckHealthUseCase(nil)

	_, err := useCase.Execute(context.Background())
	if !errors.Is(err, inport.ErrClockNotConfigured) {
		t.Fatalf("unexpected error: got %v", err)
	}
}
