package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"time"

	"payrune/internal/adapters/inbound/http/middleware"
	"payrune/internal/infrastructure/di"
)

func Run(ctx context.Context, addr string) error {
	container, err := di.NewContainer()
	if err != nil {
		return err
	}
	defer func() {
		_ = container.Close()
	}()

	mux := http.NewServeMux()
	container.HealthController.RegisterRoutes(mux)
	container.ChainAddressController.RegisterRoutes(mux)
	handler := middleware.CORS(mux)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	err = httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
