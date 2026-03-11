package cloudflareworker

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

type Request struct {
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	RawQuery string            `json:"rawQuery"`
	Headers  map[string]string `json:"headers"`
	Body     string            `json:"body"`
}

type Response struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Adapter struct {
	handler http.Handler
}

func NewAdapter(handler http.Handler) *Adapter {
	return &Adapter{handler: handler}
}

func (a *Adapter) Handle(ctx context.Context, request Request) (Response, error) {
	if a == nil || a.handler == nil {
		return Response{}, errors.New("cloudflare worker handler is not configured")
	}

	method := strings.TrimSpace(request.Method)
	if method == "" {
		method = http.MethodGet
	}
	path := strings.TrimSpace(request.Path)
	if path == "" {
		path = "/"
	}

	targetURL := &url.URL{
		Scheme:   "https",
		Host:     "worker.local",
		Path:     path,
		RawQuery: strings.TrimSpace(request.RawQuery),
	}
	httpRequest, err := http.NewRequestWithContext(ctx, method, targetURL.String(), strings.NewReader(request.Body))
	if err != nil {
		return Response{}, err
	}
	for name, value := range request.Headers {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue
		}
		httpRequest.Header.Set(trimmedName, value)
	}

	recorder := httptest.NewRecorder()
	a.handler.ServeHTTP(recorder, httpRequest)

	result := recorder.Result()
	defer func() {
		_ = result.Body.Close()
	}()
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return Response{}, err
	}

	headers := make(map[string]string, len(result.Header))
	for name, values := range result.Header {
		if len(values) == 0 {
			continue
		}
		headers[name] = strings.Join(values, ", ")
	}

	return Response{
		Status:  result.StatusCode,
		Headers: headers,
		Body:    string(bodyBytes),
	}, nil
}
