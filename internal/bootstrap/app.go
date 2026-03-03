package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"time"

	"payrune/internal/infrastructure/di"
)

func Run(ctx context.Context, addr string) error {
	container := di.NewContainer()
	mux := http.NewServeMux()
	container.HealthController.RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	err := httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
