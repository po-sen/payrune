package cloudflareworker

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	httpcontroller "payrune/internal/adapters/inbound/http/controllers"
	inport "payrune/internal/application/ports/inbound"
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

type APIDependencies struct {
	CheckHealthUseCase             inport.CheckHealthUseCase
	ListAddressPoliciesUseCase     inport.ListAddressPoliciesUseCase
	GenerateAddressUseCase         inport.GenerateAddressUseCase
	AllocatePaymentAddressUseCase  inport.AllocatePaymentAddressUseCase
	GetPaymentAddressStatusUseCase inport.GetPaymentAddressStatusUseCase
}

func NewAPIHandler(deps APIDependencies) http.Handler {
	healthController := httpcontroller.NewHealthController(deps.CheckHealthUseCase)
	chainAddressController := httpcontroller.NewChainAddressController(
		deps.ListAddressPoliciesUseCase,
		deps.GenerateAddressUseCase,
		deps.AllocatePaymentAddressUseCase,
		deps.GetPaymentAddressStatusUseCase,
	)

	mux := http.NewServeMux()
	healthController.RegisterRoutes(mux)
	chainAddressController.RegisterRoutes(mux)
	return mux
}

func HandleRequest(ctx context.Context, handler http.Handler, request Request) (Response, error) {
	if handler == nil {
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
	handler.ServeHTTP(recorder, httpRequest)

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
