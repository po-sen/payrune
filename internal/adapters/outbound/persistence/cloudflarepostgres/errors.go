package cloudflarepostgres

import (
	"errors"

	cloudflarepostgresinfra "payrune/internal/infrastructure/cloudflarepostgres"
)

func isUniqueViolation(err error, constraint string) bool {
	var queryErr *cloudflarepostgresinfra.QueryError
	if !errors.As(err, &queryErr) {
		return false
	}
	return queryErr.Code == "23505" && queryErr.Constraint == constraint
}
