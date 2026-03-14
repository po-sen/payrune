package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"payrune/internal/infrastructure/di"
)

type apiWorkerRequestEnvelope struct {
	Request  apiWorkerRequest  `json:"request"`
	Env      map[string]string `json:"env"`
	BridgeID string            `json:"bridgeId"`
}

type apiWorkerResponseEnvelope struct {
	Response apiWorkerResponse `json:"response"`
}

type apiWorkerRequest struct {
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	RawQuery string            `json:"rawQuery"`
	Headers  map[string]string `json:"headers"`
	Body     string            `json:"body"`
}

type apiWorkerResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func HandleCloudflareAPIRequestJSON(ctx context.Context, payload string) (string, error) {
	var envelope apiWorkerRequestEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}

	response, err := handleCloudflareAPIRequest(ctx, envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(apiWorkerResponseEnvelope{Response: response})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleCloudflareAPIRequest(
	ctx context.Context,
	envelope apiWorkerRequestEnvelope,
) (apiWorkerResponse, error) {
	handler, err := di.BuildCloudflareAPIHTTPHandler(envelope.Env, envelope.BridgeID)
	if err != nil {
		return apiWorkerResponse{}, err
	}

	return executeAPIWorkerRequest(ctx, handler, envelope.Request)
}

func executeAPIWorkerRequest(
	ctx context.Context,
	handler http.Handler,
	request apiWorkerRequest,
) (apiWorkerResponse, error) {
	if handler == nil {
		return apiWorkerResponse{}, errors.New("cloudflare worker handler is not configured")
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
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		method,
		targetURL.String(),
		strings.NewReader(request.Body),
	)
	if err != nil {
		return apiWorkerResponse{}, err
	}
	for name, value := range request.Headers {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue
		}
		httpRequest.Header.Set(trimmedName, value)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httpRequest)

	result := recorder.Result()
	defer func() {
		_ = result.Body.Close()
	}()
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return apiWorkerResponse{}, err
	}

	headers := make(map[string]string, len(result.Header))
	for name, values := range result.Header {
		if len(values) == 0 {
			continue
		}
		headers[name] = strings.Join(values, ", ")
	}

	return apiWorkerResponse{
		Status:  result.StatusCode,
		Headers: headers,
		Body:    string(bodyBytes),
	}, nil
}
