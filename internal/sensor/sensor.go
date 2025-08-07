// Package sensor provides simulation logic for IoT sensors.
// Each sensor runs as an independent goroutine that periodically emits data to a shared channel.
// Sensors recover from panics and restart automatically.
package sensor

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
)

// Sensor encapsulates the logic for a single simulated sensor.
type Sensor struct {
	ID       int
	DataCh   chan<- model.SensorData
	Interval time.Duration
	rand     *rand.Rand
	randMux  sync.Mutex
}

// NewSensor creates and returns a new Sensor instance.
func NewSensor(id int, dataCh chan<- model.SensorData, interval time.Duration) *Sensor {
	randSrc := rand.NewSource(time.Now().UnixNano() + int64(id)) // Add the id to ensure sensors created at the exact same nanosecond have different random sequences.
	return &Sensor{
		ID:       id,
		DataCh:   dataCh,
		Interval: interval,
		rand:     rand.New(randSrc),
	}
}

// Run starts the sensor's data generation loop.
// It emits generated data to the sensors DataCh at every Interval.
// It stops when the context ctx is cancelled.
func (s *Sensor) Run(ctx context.Context) {
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	log.Printf("Sensor %d starting\n", s.ID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Sensor %d stopping\n", s.ID)
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
		}
	}
}

// Start launches a simulated sensor (identified by ID) as a goroutine with panic recovery.
// The goroutine runs the Sensor's Run method.
func Start(ctx context.Context, id int, dataCh chan<- model.SensorData, interval time.Duration) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Sensor %d panicked: %v - restarting\n", id, r)
				// Restart the sensor only if the context is not done.
				// This prevents a panic-restart loop if the context is cancelled.
				if ctx.Err() == nil {
					Start(ctx, id, dataCh, interval)
				}
			}
		}()

		s := NewSensor(id, dataCh, interval)
		s.Run(ctx)
	}()
}
