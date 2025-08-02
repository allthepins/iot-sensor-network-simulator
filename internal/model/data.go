// Package model defines shared data structures used across the simulator.
package model

import "time"

// SensorData represents a single reading emitted by a simulated sensor.
type SensorData struct {
	ID        int
	Value     float64
	Timestamp time.Time
}
