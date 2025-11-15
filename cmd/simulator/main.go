// Package main starts the IoT Sensor Network Simulator.
// It configures the sensor network, runs the aggregator,
// and ensures the graceful shutdown of all components.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/allthepins/iot-sensor-network-simulator/internal/aggregator"
	"github.com/allthepins/iot-sensor-network-simulator/internal/logging"
	"github.com/allthepins/iot-sensor-network-simulator/internal/metrics"
	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
	"github.com/allthepins/iot-sensor-network-simulator/internal/nats"
	"github.com/allthepins/iot-sensor-network-simulator/internal/publisher"
	"github.com/allthepins/iot-sensor-network-simulator/internal/sensor"
	"github.com/allthepins/iot-sensor-network-simulator/internal/server"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	// Simulation and metrics parameters
	// TODO Set simulation params via args or config values
	var (
		sensorCount        = 5000
		simulationDuration = 10 * time.Minute // Increased simulation duration to allow more time to monitor metrics.
		sensorInterval     = 100 * time.Millisecond
		metricsAddr        = ":2112"
		pprofAddr          = ":6060"
		enableNATS         = true // Feature flag for NATS integration. TODO Set via env var
	)

	// logging setup
	logger := logging.NewJSONLogger()
	slog.SetDefault(logger)

	// Metrics and Server setup
	reg := prometheus.NewRegistry()
	appMetrics := metrics.NewMetrics(reg)
	metricsServer := server.NewMetricsServer(metricsAddr, reg)

	// Main context that can be cancelled by an OS signal (e.g `ctrl+c`).
	mainCtx, stopMain := context.WithCancel(context.Background())

	// Start the metrics server in a separate goroutine.
	go metricsServer.Serve(mainCtx)

	// Start the pprof server in a separate goroutine.
	// This allows us to use go pprof tool profiling.
	go server.StartPprofServer(mainCtx, pprofAddr)

	// NATS setup (`enableNATS` feature flag controlled)
	var natsClient *nats.Client
	var publisherWg sync.WaitGroup

	if enableNATS {
		natsURL := os.Getenv("NATS_URL")
		if natsURL == "" {
			natsURL = "nats://localhost:4222"
		}

		natsCfg := nats.DefaultConfig()
		natsCfg.URL = natsURL

		var err error
		natsClient, err = nats.NewClient(natsCfg, logger)
		if err != nil {
			logger.Error("Failed to connect to NATS, continuiong without NATS", "error", err)
			appMetrics.NATSConnectionStatus.Set(0)
			enableNATS = false
		} else {
			logger.Info("NATS client initialized", "url", natsURL)
			appMetrics.NATSConnectionStatus.Set(1)

			defer func() {
				if err := natsClient.Close(); err != nil {
					logger.Error("Error closing NATS client", "error", err)
				}
			}()
		}
	}

	// Channel to listen for interrupt signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt) // Listen for SIGINT

	// Launch a goroutine to wait for a SIGINT signal.
	// It cancels the main context if it receives one.
	go func() {
		<-sigCh
		logger.Info("Shutdown signal received, starting graceful shutdown.")
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
		aggregator.New(dataCh, appMetrics, logger).Run(ctx)
	}()

	// Start the NATS publisher.
	if enableNATS && natsClient != nil {
		publisherWg.Add(1)
		go func() {
			defer publisherWg.Done()

			pub := publisher.New(dataCh, natsClient, nats.DefaultSubjectPrefix, appMetrics, logger)
			pub.Run(ctx)
		}()

		// Periodically check and update NATS connection status
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if natsClient.IsConnected() {
						appMetrics.NATSConnectionStatus.Set(1)
					} else {
						appMetrics.NATSConnectionStatus.Set(0)
					}
				}
			}
		}()
	}

	// Start sensors.
	for i := 1; i <= sensorCount; i++ {
		sensorsWg.Add(1)

		// TODO Look into refactoring `sensor.Start` such that we can directly wait for it,
		// rather than having to wrap its invocation in another goroutine (so it can be integrated with sensorsWg).
		go func(id int, interval time.Duration) {
			defer sensorsWg.Done()

			sensor.Start(ctx, id, dataCh, interval, appMetrics, logger)
			// Wait for the shutdown signal from the context.
			// When the context is cancelled, the sensor's internal goroutine alse receives the signal and will terminate.
			// This ensures Done() is called only after the sensor is asked to stop,
			<-ctx.Done()
		}(i, sensorInterval)
	}

	logger.Info("Simulation starting",
		"sensor_count", sensorCount,
		"simulation_duration", simulationDuration,
		"nats_enabled", enableNATS,
	)

	// Launch a dedicated goroutine to orchestrate the shutdown of sensors.
	go func() {
		// Wait for sensors to be done.
		// (When their context is cancelled or the simulationDuration elapses).
		sensorsWg.Wait()

		// Now safe to close the data channel.
		close(dataCh)
		logger.Info("All sensors shutdown. Data channel closed.")
	}()

	// Wait for the aggregator.
	aggregatorWg.Wait()

	// Wait for the NATS publisher.
	if enableNATS {
		publisherWg.Wait()
		logger.Info("NATS publisher shutdown complete.")
	}

	logger.Info("Simulation ended gracefully.")
}
