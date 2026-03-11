package cloudflarepostgres

import (
	"fmt"
	"strings"
	"time"
)

func nullIfEmpty(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableTimePointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

func buildSequentialPlaceholders(start int, count int) string {
	if count <= 0 {
		return ""
	}
	parts := make([]string, 0, count)
	for i := 0; i < count; i++ {
		parts = append(parts, fmt.Sprintf("$%d", start+i))
	}
	return strings.Join(parts, ", ")
}
