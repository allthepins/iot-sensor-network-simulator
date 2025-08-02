// Package aggregator receives and processes data from all active sensors.
// It runs as a single goroutine, reading from a shared channel.
package aggregator

import (
	"log"

	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
)

// Run starts the aggregator loop.
// It reads SensorData from dataCh and processes it.
// Once dataCh is closed, Run signals completion via doneCh.
func Run(dataCh <-chan model.SensorData, doneCh chan<- struct{}) {
	for data := range dataCh {
		log.Printf("Aggregator received: Sensor %d - %f", data.ID, data.Value)
	}
	doneCh <- struct{}{}
}
