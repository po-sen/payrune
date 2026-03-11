package cloudflarepostgres

import "errors"

type QueryError struct {
	Message    string
	Code       string
	Constraint string
}

func (e *QueryError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return "postgres query failed"
}

func isUniqueViolation(err error, constraint string) bool {
	var queryErr *QueryError
	if !errors.As(err, &queryErr) {
		return false
	}
	return queryErr.Code == "23505" && queryErr.Constraint == constraint
}
