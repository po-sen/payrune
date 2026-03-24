package cloudflarepostgres

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
