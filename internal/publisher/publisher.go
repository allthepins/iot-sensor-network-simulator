// Package publisher provides functionality for
// publishing sensor data from a Go channel to NATS.
package publisher

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/metrics"
	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
	"github.com/allthepins/iot-sensor-network-simulator/internal/nats"
)

// Publisher reads sensor data from a channel and publishes it to NATS.
type Publisher struct {
	dataCh        <-chan model.SensorData
	natsClient    *nats.Client
	subjectPrefix string
	metrics       *metrics.Metrics
	logger        *slog.Logger
}

// New creates a new Publisher instance.
func New(dataCh <-chan model.SensorData, natsClient *nats.Client, subjectPrefix string, m *metrics.Metrics, l *slog.Logger) *Publisher {
	if l == nil {
		l = slog.Default()
	}

	return &Publisher{
		dataCh:        dataCh,
		natsClient:    natsClient,
		subjectPrefix: subjectPrefix,
		metrics:       m,
		logger:        l.With("component", "publisher"),
	}
}

// Run starts the publisher loop (that reads from the data channel and pulishes to NATS).
// It continues until the context is canceled or the data channel is closed.
func (p *Publisher) Run(ctx context.Context) {
	p.logger.Info("Publisher starting")
	defer p.logger.Info("Publisher stopping")

	// ticker to trigger periodic logging of publish statistics
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	successCount := 0
	failureCount := 0

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Publisher context canceled",
				"success", successCount,
				"failures", failureCount)
			return

		case data, ok := <-p.dataCh:
			if !ok {
				p.logger.Info("Data channel closed",
					"success", successCount,
					"failures", failureCount)
				return
			}

			if err := p.publish(ctx, data); err != nil {
				p.logger.Warn("Failed to publish to NATS",
					"sensor_id", data.ID,
					"error", err)
				failureCount++

				if p.metrics != nil {
					p.metrics.NATSPublishFailures.WithLabelValues(
						strconv.Itoa(data.ID),
						"publish_error",
					).Inc()
				}
			} else {
				successCount++

				if p.metrics != nil {
					p.metrics.NATSPublishSuccess.WithLabelValues(
						strconv.Itoa(data.ID),
					).Inc()
				}
			}

		case <-ticker.C:
			p.logger.Info("Publisher statistics",
				"success", successCount,
				"failures", failureCount,
				"nats_connected", p.natsClient.IsConnected(),
			)
		}
	}
}

// publish publishes a single SensorData message to NATS.
func (p *Publisher) publish(ctx context.Context, data model.SensorData) error {
	if !p.natsClient.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	// Construct the message subject as `iot.sensors.data.{sensor_id}`
	subject := fmt.Sprintf("%s.data.%d", p.subjectPrefix, data.ID)

	// Measure publish latency
	start := time.Now()

	publishCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := p.natsClient.PublishJson(publishCtx, subject, data)

	if p.metrics != nil {
		duration := time.Since(start).Seconds()
		p.metrics.NATSPublishLatency.WithLabelValues(
			strconv.Itoa(data.ID),
		).Observe(duration)
	}

	return err
}
