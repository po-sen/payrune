package controllers

import (
	"log"
	"net/http"
)

func logMappedControllerError(
	r *http.Request,
	statusCode int,
	publicMessage string,
	err error,
) {
	if r == nil || err == nil {
		return
	}

	log.Printf(
		"api request failed method=%s path=%s status=%d public_error=%q err=%v",
		r.Method,
		r.URL.Path,
		statusCode,
		publicMessage,
		err,
	)
}
