// Package sensor provides simulation logic for IoT sensors.
// Each sensor runs as an independent goroutine that periodically emits data to a shared channel.
// Sensors recover from panics and restart automatically.
package sensor

import (
	"log"
	"math/rand"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
)

// Start launches a simulated sensor (identified by ID) as a goroutine.
// It periodically emits SensorData to the provided dataCh.
// The sensor stops when stopCh is closed.
// If the goroutine panics, it is automatically restarted.
func Start(id int, dataCh chan<- model.SensorData, stopCh <-chan struct{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Sensor %d panicked: %v - restarting\n", id, r)
				Start(id, dataCh, stopCh)
			}
		}()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				log.Printf("Sensor %d stopping\n", id)
				return
			case <-ticker.C:
				data := model.SensorData{
					ID:        id,
					Value:     rand.Float64(),
					Timestamp: time.Now(),
				}
				dataCh <- data
			}
		}
	}()
}
