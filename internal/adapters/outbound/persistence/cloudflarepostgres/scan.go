package cloudflarepostgres

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"time"
)

func scanValues(values []any, dest ...any) error {
	if len(values) != len(dest) {
		return fmt.Errorf("scan destination mismatch: got %d values for %d destinations", len(values), len(dest))
	}

	for i := range dest {
		if err := assignValue(dest[i], values[i]); err != nil {
			return fmt.Errorf("scan column %d: %w", i, err)
		}
	}
	return nil
}

func assignValue(dest any, value any) error {
	switch target := dest.(type) {
	case *string:
		if target == nil {
			return fmt.Errorf("destination string pointer is nil")
		}
		if value == nil {
			*target = ""
			return nil
		}
		text, err := asString(value)
		if err != nil {
			return err
		}
		*target = text
		return nil
	case *int64:
		if target == nil {
			return fmt.Errorf("destination int64 pointer is nil")
		}
		number, err := asInt64(value)
		if err != nil {
			return err
		}
		*target = number
		return nil
	case *int32:
		if target == nil {
			return fmt.Errorf("destination int32 pointer is nil")
		}
		number, err := asInt64(value)
		if err != nil {
			return err
		}
		*target = int32(number)
		return nil
	case *sql.NullString:
		if target == nil {
			return fmt.Errorf("destination sql.NullString pointer is nil")
		}
		if value == nil {
			*target = sql.NullString{}
			return nil
		}
		text, err := asString(value)
		if err != nil {
			return err
		}
		*target = sql.NullString{String: text, Valid: true}
		return nil
	case *sql.NullInt64:
		if target == nil {
			return fmt.Errorf("destination sql.NullInt64 pointer is nil")
		}
		if value == nil {
			*target = sql.NullInt64{}
			return nil
		}
		number, err := asInt64(value)
		if err != nil {
			return err
		}
		*target = sql.NullInt64{Int64: number, Valid: true}
		return nil
	case *sql.NullInt32:
		if target == nil {
			return fmt.Errorf("destination sql.NullInt32 pointer is nil")
		}
		if value == nil {
			*target = sql.NullInt32{}
			return nil
		}
		number, err := asInt64(value)
		if err != nil {
			return err
		}
		*target = sql.NullInt32{Int32: int32(number), Valid: true}
		return nil
	case *sql.NullTime:
		if target == nil {
			return fmt.Errorf("destination sql.NullTime pointer is nil")
		}
		if value == nil {
			*target = sql.NullTime{}
			return nil
		}
		timestamp, err := asTime(value)
		if err != nil {
			return err
		}
		*target = sql.NullTime{Time: timestamp, Valid: true}
		return nil
	default:
		return fmt.Errorf("unsupported scan destination type %T", dest)
	}
}

func asString(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []byte:
		return string(typed), nil
	case int:
		return strconv.Itoa(typed), nil
	case int32:
		return strconv.FormatInt(int64(typed), 10), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case float64:
		if math.Trunc(typed) == typed {
			return strconv.FormatInt(int64(typed), 10), nil
		}
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case bool:
		if typed {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unsupported string source type %T", value)
	}
}

func asInt64(value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int32:
		return int64(typed), nil
	case int64:
		return typed, nil
	case float64:
		if math.Trunc(typed) != typed {
			return 0, fmt.Errorf("non-integer numeric value %v", typed)
		}
		return int64(typed), nil
	case string:
		number, err := strconv.ParseInt(typed, 10, 64)
		if err != nil {
			return 0, err
		}
		return number, nil
	default:
		return 0, fmt.Errorf("unsupported int64 source type %T", value)
	}
}

func asTime(value any) (time.Time, error) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), nil
	case string:
		formats := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05.999999999Z07:00",
			"2006-01-02 15:04:05Z07:00",
		}
		for _, format := range formats {
			if parsed, err := time.Parse(format, typed); err == nil {
				return parsed.UTC(), nil
			}
		}
		return time.Time{}, fmt.Errorf("unsupported time value %q", typed)
	default:
		return time.Time{}, fmt.Errorf("unsupported time source type %T", value)
	}
}
