// Package server provides functionality for running an HTTP server.
package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer is an HTTP server for exposing Prometheus metrics.
type MetricsServer struct {
	server *http.Server
}

// NewMetricsServer creates a new MetricsServer.
// It accepts an address addr (e.g. ":2112") and a Prometheus registry reg.
func NewMetricsServer(addr string, reg *prometheus.Registry) *MetricsServer {
	mux := http.NewServeMux()
	// Create a new handler for the given registry.
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	mux.Handle("/metrics", promHandler)

	return &MetricsServer{
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

// Serve starts the HTTP server and handles graceful shutdown.
func (s *MetricsServer) Serve(ctx context.Context) {
	go func() {
		log.Printf("Metrics server starting on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ERROR: Metrics server failed: %v", err) // NOTE Exit out of entire application. Might make sense to change this eventually.
		}
	}()

	// Wait for the context to be done, which signals shutdown.
	<-ctx.Done()
	log.Println("Shutting down metrics server...")

	// Create a context with a timeout for the shutdown process.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("ERROR: Metrics server shutdown failed: %v", err)
	}
}
