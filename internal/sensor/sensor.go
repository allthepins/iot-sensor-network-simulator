// Package sensor provides simulation logic for IoT sensors.
// Each sensor runs as an independent goroutine that periodically emits data to a shared channel.
// Sensors recover from panics and restart automatically.
package sensor

import (
	"context"
	"log/slog"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/metrics"
	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
)

// Sensor encapsulates the logic for a single simulated sensor.
type Sensor struct {
	ID       int
	DataCh   chan<- model.SensorData
	Interval time.Duration
	rand     *rand.Rand
	randMux  sync.Mutex
	idStr    string // Store ID as a string for performance when labeling metrics.
	metrics  *metrics.Metrics
	logger   *slog.Logger
}

// NewSensor creates and returns a new Sensor instance.
func NewSensor(id int, dataCh chan<- model.SensorData, interval time.Duration, m *metrics.Metrics, l *slog.Logger) *Sensor {
	if l == nil {
		l = slog.Default()
	}

	randSrc := rand.NewSource(time.Now().UnixNano() + int64(id)) // Add the id to ensure sensors created at the exact same nanosecond have different random sequences.
	return &Sensor{
		ID:       id,
		DataCh:   dataCh,
		Interval: interval,
		rand:     rand.New(randSrc),
		idStr:    strconv.Itoa(id), // Convert ID to string once.
		metrics:  m,
		logger:   l.With("component", "sensor", "sensor_id", id),
	}
}

// Run starts the sensor's data generation loop.
// It emits generated data to the sensors DataCh at every Interval.
// It stops when the context ctx is cancelled.
func (s *Sensor) Run(ctx context.Context) {
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	s.logger.Info("Sensor starting", "sensor_id", s.ID)

	if s.metrics != nil {
		s.metrics.ActiveSensors.Inc()
		defer s.metrics.ActiveSensors.Dec()
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Sensor stopping", "sensor_id", s.ID)
			return
		case <-ticker.C:
			// Use a mutex to make random number generation safe for concurrent access
			s.randMux.Lock()
			value := s.rand.Float64()
			s.randMux.Unlock()

			data := model.SensorData{
				ID:        s.ID,
				Value:     value,
				Timestamp: time.Now(),
			}
			s.DataCh <- data

			// Instrument the message send and value observation.
			if s.metrics != nil {
				s.metrics.MessagesSent.WithLabelValues(s.idStr).Inc()
				s.metrics.GeneratedValues.WithLabelValues(s.idStr).Observe(value)
			}
		}
	}
}

// Start launches a simulated sensor (identified by ID) as a goroutine with panic recovery.
// The goroutine runs the Sensor's Run method.
func Start(ctx context.Context, id int, dataCh chan<- model.SensorData, interval time.Duration, m *metrics.Metrics, l *slog.Logger) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicLogger := l.With("component", "sensor", "sensor_id", id)
				panicLogger.Error("Sensor panicked - restarting", "sensor_id", id, "panic", r)

				// Restart the sensor only if the context is not done.
				// This prevents a panic-restart loop if the context is cancelled.
				if ctx.Err() == nil {
					// Instrument the restart.
					if m != nil {
						m.SensorRestarts.WithLabelValues(strconv.Itoa(id)).Inc()
					}

					Start(ctx, id, dataCh, interval, m, l)
				}
			}
		}()

		s := NewSensor(id, dataCh, interval, m, l)
		s.Run(ctx)
	}()
}
