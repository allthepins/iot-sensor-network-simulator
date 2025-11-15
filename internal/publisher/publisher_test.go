// Package publisher_test contains tests for the publisher package.
package publisher_test

import (
	"context"
	"testing"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
	"github.com/allthepins/iot-sensor-network-simulator/internal/publisher"
)

// TestNew verifies that New creates a Publisher instance.
func TestNew(t *testing.T) {
	t.Parallel()

	dataCh := make(chan model.SensorData)
	pub := publisher.New(dataCh, nil, "iot.sensors", nil, nil)

	if pub == nil {
		t.Fatal("New returned nil")
	}
}

// TestPublisher_Run_StopsOnContextCancel verifies the publisher stops when context is canceled.
func TestPublisher_Run_StopsOnContextCancel(t *testing.T) {
	t.Parallel()

	dataCh := make(chan model.SensorData)
	pub := publisher.New(dataCh, nil, "iot.sensors", nil, nil)

	ctx, cancel := context.WithCancel(context.Background())

	runFinished := make(chan struct{})
	go func() {
		pub.Run(ctx)
		close(runFinished)
	}()

	// Cancel context immediately
	cancel()

	select {
	case <-runFinished:
		// Expected behavior: Run exited gracefully
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Publisher did not stop after context cancellation")
	}
}

// TestPublisher_Run_StopsOnChannelClose verifies the publisher stops when data channel is closed.
func TestPublisher_Run_StopsOnChannelClose(t *testing.T) {
	t.Parallel()

	dataCh := make(chan model.SensorData)
	pub := publisher.New(dataCh, nil, "iot.sensors", nil, nil)

	ctx := context.Background()

	runFinished := make(chan struct{})
	go func() {
		pub.Run(ctx)
		close(runFinished)
	}()

	// Close the data channel
	close(dataCh)

	select {
	case <-runFinished:
		// Expected behavior: Run exited gracefully
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Publisher did not stop after channel close")
	}
}

// TODO: Integration tests with a real NATS connection:
// - successful publishing to NATS
// - error handling when NATS is unavailable
// - metrics recording
// - publishing multiple messages
// - subject formatting
