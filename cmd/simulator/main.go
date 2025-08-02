// Package main starts the IoT Sensor Network Simulator.
// It configures the sensor network, runs the aggregator,
// and ensures graceful shutdown of all components.
package main

import (
	"fmt"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/aggregator"
	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
	"github.com/allthepins/iot-sensor-network-simulator/internal/sensor"
)

func main() {
	// Number of sensors
	// TODO Set this via arg or config value
	const sensorCount = 5000

	// Number of seconds simulation should run for (in seconds)
	// TODO Set this via arg or config value
	const simulationDuration = 10

	// Channels
	dataCh := make(chan model.SensorData, 1000) // Buffered channel sensors send data to.
	stopCh := make(chan struct{})               // Channel we signal to stop sensors.
	done := make(chan struct{})                 // Channel the aggregator signals when it's done.

	// Start aggregator
	go aggregator.Run(dataCh, done)

	// Start sensors
	for i := 1; i <= sensorCount; i++ {
		go sensor.Start(i, dataCh, stopCh)
	}

	// Let simulation run for given duration
	time.Sleep(simulationDuration * time.Second)

	// Signal sensors to stop
	close(stopCh)

	// Close data channel once sensors have stopped sending
	// TODO Switch to sync.Waitgroup to wait for sensors
	time.Sleep(1 * time.Second)
	close(dataCh)

	// Wait for aggregator to signal it's done
	<-done
	fmt.Println("Simulation ended gracefully")
}
