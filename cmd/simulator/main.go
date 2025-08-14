// Package main starts the IoT Sensor Network Simulator.
// It configures the sensor network, runs the aggregator,
// and ensures graceful shutdown of all components.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/allthepins/iot-sensor-network-simulator/internal/aggregator"
	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
	"github.com/allthepins/iot-sensor-network-simulator/internal/sensor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// startMetricsServer starts the Prometheus metrics server.
func startMetricsServer(ctx context.Context) {
	server := &http.Server{Addr: ":2112"}
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Println("Metrics server starting on :2112")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Metrics server failed: %v", err)
		}
	}()

	// Wait for the main context to be cancelled then gracefully shutdown the server.
	<-ctx.Done()
	log.Println("Shutting down metrics server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Metrics server shutdown failed: %v", err)
	}
}

// startPprofServer starts an HTTP server with pprof enabled on :6060/debug/pprof/
func startPprofServer(ctx context.Context) {
	server := &http.Server{
		Addr:    ":6060",
		Handler: http.DefaultServeMux, // Uses the default mux where pprof is registered
	}

	go func() {
		log.Println("pprof server starting on :6060")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("pprof server failed: %v", err)
		}
	}()

	// Graceful shutdown on context cancellation
	<-ctx.Done()
	log.Println("Shutting down pprof server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("pprof server shutdown failed: %v", err)
	}
}

func main() {
	// Simulation parameters
	// TODO Set these via args or config values
	var (
		sensorCount        = 5000
		simulationDuration = 10 * time.Minute // Increased simulation duration to allow more time to monitor metrics.
		sensorInterval     = 100 * time.Millisecond
	)

	// Main context that can be cancelled by an OS signal (e.g `ctrl+c`).
	mainCtx, stopMain := context.WithCancel(context.Background())

	// Start the metrics server in a separate goroutine.
	go startMetricsServer(mainCtx)

	// Start the pprof server in a separate goroutine.
	// This allows us to use go pprof tool profiling.
	go startPprofServer(mainCtx)

	// Channel to listen for interrupt signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt) // Listen for SIGINT

	// Launch a goroutine to wait for a SIGINT signal.
	// It cancels the main context if it receives one.
	go func() {
		<-sigCh
		log.Println("Shutdown signal received, starting graceful shutdown.")
		stopMain()
	}()

	// Create a derived context that is automatically cancelled after the simulation duration,
	// or by the main context being cancelled by an OS interrupt.
	// This context is the primary signal for all goroutines to begin graceful shutdown.
	ctx, cancel := context.WithTimeout(mainCtx, simulationDuration)
	defer cancel()

	// Buffered channel sensors send data to.
	dataCh := make(chan model.SensorData, 1000)

	// WaitGroups to coordinate a graceful shutdown.
	// sensorsWg for the sensors.
	// aggregatorWg for the aggregator.
	var sensorsWg, aggregatorWg sync.WaitGroup

	// Start the aggregator.
	aggregatorWg.Add(1)
	go func() {
		defer aggregatorWg.Done()

		// Instantiate and run the aggregator.
		// It should run until its context is cancelled
		// and the data channel is drained and closed.
		aggregator.New(dataCh).Run(ctx)
	}()

	// Start sensors.
	for i := 1; i <= sensorCount; i++ {
		sensorsWg.Add(1)

		// TODO Look into refactoring `sensor.Start` such that we can directly wait for it,
		// rather than having to wrap its invocation in another goroutine (so it can be integrated with sensorsWg).
		go func(id int, interval time.Duration) {
			defer sensorsWg.Done()

			sensor.Start(ctx, id, dataCh, interval)
			// Wait for the shutdown signal from the context.
			// When the context is cancelled, the sensor's internal goroutine alse receives the signal and will terminate.
			// This ensures Done() is called only after the sensor is asked to stop,
			<-ctx.Done()
		}(i, sensorInterval)
	}

	log.Printf("Simulation starting with %d sensors for %s.", sensorCount, simulationDuration)

	// Launch a dedicated goroutine to orchestrate the shutdown of sensors.
	go func() {
		// Wait for sensors to be done.
		// (When their context is cancelled or the simulationDuration elapses).
		sensorsWg.Wait()

		// Now safe to close the data channel.
		close(dataCh)
		log.Println("All sensors shutdown. Data channel closed.")
	}()

	// Wait for the aggregator.
	aggregatorWg.Wait()

	log.Println("Simulation ended gracefully.")
}
