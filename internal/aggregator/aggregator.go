// Package aggregator receives and processes data from all active sensors.
// It runs as a single goroutine, reading from a shared channel until its context is canceled.
package aggregator

import (
	"context"
	"log/slog"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/metrics"
	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
)

// Aggregator processes sensor data.
type Aggregator struct {
	DataCh  <-chan model.SensorData
	metrics *metrics.Metrics
	logger  *slog.Logger
}

// New creates and returns a new Aggregator instance.
func New(dataCh <-chan model.SensorData, m *metrics.Metrics, l *slog.Logger) *Aggregator {
	if l == nil {
		l = slog.Default() // Fallback to default logger if nil logger provided.
	}

	return &Aggregator{
		DataCh:  dataCh,
		metrics: m,
		logger:  l.With("component", "aggregator"),
	}
}

// Run starts the aggregator loop, which reads and processes SensorData.
// It listens for data on its DataCh and processes it.
// The loop terminates when the given context is canceled, or if DataCh is closed.
func (a *Aggregator) Run(ctx context.Context) {
	a.logger.Info("Aggregator starting")
	defer a.logger.Info("Aggregator stopping")

	// Use a ticker and counter to help log a summary of processed messages every 5 seconds.
	summaryTicker := time.NewTicker(5 * time.Second)
	defer summaryTicker.Stop()
	count := 0

	for {
		select {
		case <-ctx.Done():
			// Context has been canceled, so we exit.
			return
		case _, ok := <-a.DataCh:
			// The `ok` flag is false if DataCh has been closed.
			if !ok {
				return
			}

			// Instrument the message receipt.
			if a.metrics != nil {
				a.metrics.MessagesReceived.Inc()
			}

			count++
		case <-summaryTicker.C:
			a.logger.Info("processed messages", "count", count)
		}
	}
}
