// Package aggregator receives and processes data from all active sensors.
// It runs as a single goroutine, reading from a shared channel until its context is canceled.
package aggregator

import (
	"context"
	"log"

	"github.com/allthepins/iot-sensor-network-simulator/internal/metrics"
	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
)

// Aggregator processes sensor data.
type Aggregator struct {
	DataCh <-chan model.SensorData
}

// New creates and returns a new Aggregator instance.
func New(dataCh <-chan model.SensorData) *Aggregator {
	return &Aggregator{
		DataCh: dataCh,
	}
}

// Run starts the aggregator loop, which reads and processes SensorData.
// It listens for data on its DataCh and processes it.
// The loop terminates when the given context is canceled, or if DataCh is closed.
func (a *Aggregator) Run(ctx context.Context) {
	log.Println("Aggregator starting")
	defer log.Println("Aggregator stopping")

	for {
		select {
		case <-ctx.Done():
			// Context has been canceled, so we exit.
			return
		case data, ok := <-a.DataCh:
			// The `ok` flag is false if DataCh has been closed.
			if !ok {
				return
			}

			// Instrument the message receipt.
			metrics.MessagesReceived.Inc()

			log.Printf("Aggregator received: Sensor %d - %f", data.ID, data.Value)
		}
	}
}
