package server

import (
	"context"
	"log"
	"net/http"
	"net/http/pprof"
	"time"
)

// StartPprofServer starts a dedicated HTTP server for pprof profiling endpoints.
func StartPprofServer(ctx context.Context, addr string) {
	mux := http.NewServeMux()

	// Explicitly register the pprof handlers.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("pprof server starting on %s", addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("ERROR: pprof server failed: %v", err)
		}
	}()

	// Wait for the context to be cancelled to start graceful shutdown.
	<-ctx.Done()

	log.Println("Shutting down pprof server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("ERROR: pprof server shutdown failed: %v", err)
	}
}
