// Package sensor_test contains tests for the sensor package.
package sensor_test

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
	"github.com/allthepins/iot-sensor-network-simulator/internal/sensor"
)

// TestNewSensor verifies that the NewSensor function correctly initializes a Sensor.
func TestNewSensor(t *testing.T) {
	t.Parallel()

	id := 1
	interval := 100 * time.Millisecond
	dataCh := make(chan model.SensorData)

	s := sensor.NewSensor(id, dataCh, interval)

	if s == nil {
		t.Fatal("NewSensor returned nil")
	}
	if s.ID != id {
		t.Errorf("expected sensor ID %d, got %d", id, s.ID)
	}
	if s.Interval != interval {
		t.Errorf("expecteds interval %v, got %v", interval, s.Interval)
	}
	if s.DataCh == nil {
		t.Error("DataCh was not initialized")
	}
}

// TestSensor_Run tests the Sensor's Run method.
// It tests data emission and expected behavior upon context cancellation.
func TestSensor_Run(t *testing.T) {
	t.Parallel()

	interval := 10 * time.Millisecond
	dataCh := make(chan model.SensorData, 1) // Buffered channel to prevent blocking
	s := sensor.NewSensor(1, dataCh, interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Run(ctx)
	}()

	// Verify data is sent to data channel.
	select {
	case data := <-dataCh:
		if data.ID != s.ID {
			t.Errorf("expected data from sensor ID %d, got %d", s.ID, data.ID)
		}
		if data.Value < 0 || data.Value > 1 {
			t.Errorf("expected value between 0 and 1, got %f", data.Value)
		}
	case <-time.After(interval * 2):
		t.Fatal("timed out waiting for sensor data")
	}

	// Verify the sensor stops when the context is canceled.
	cancel()
	wg.Wait() // Wait for the Run method to return.

	// Verify no more data is sent after stopping.
	select {
	case d := <-dataCh:
		t.Errorf("received data after context was canceled: %+v", d)
	case <-time.After(interval * 2):
		// Expected outcome: nothing happens.
	}
}

// TestStart verifies that the Start function launches a sensor goroutine
// that sends data to a data channel and can be stopped.
func TestStart(t *testing.T) {
	t.Parallel()

	interval := 10 * time.Millisecond
	dataCh := make(chan model.SensorData, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sensor.Start(ctx, 1, dataCh, interval)

	// Verify data is being sent.
	select {
	case <-dataCh:
	// Expected behavior: data received.
	case <-time.After(interval * 2):
		t.Fatal("timed out waiting for sensor data from Start")
	}

	// Cancel the conext to stop the sensor.
	cancel()

	// Short delay to allow the gorouting to process cancellation.
	time.Sleep(interval * 2)

	// Verify no more data is sent.
	select {
	case d := <-dataCh:
		t.Errorf("received data after context cancellation: %+v", d)
	default:
		// Expected behavior: no data received.
	}
}

// TestStart_PanicRecovery verifies that a sensor goroutine will restart after a panic.
// It relies on a side-effect of the sensor restart, which is the "panicked ... restarting" log message.
// It redirects the log output to a buffer and checks it for the expected message.
// TODO Can panic recovery be tested without relying on side-effects?
func TestStart_PanicRecovery(t *testing.T) {
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr) // Restore the original logger.

	interval := 10 * time.Millisecond
	// Use a closed channel to trigger a panic when the sensor tries to send data.
	dataCh := make(chan model.SensorData)
	close(dataCh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the sensor. It should panic, recover, log, and restart in a loop.
	sensor.Start(ctx, 99, dataCh, interval)

	// Poll the log buffer for the expected panic message.
	const pollTimeout = 100 * time.Millisecond
	deadline := time.Now().Add(pollTimeout)
	for {
		if strings.Contains(logBuf.String(), "panicked: send on closed channel - restarting") {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for panic log message. Log content:\n%s", logBuf.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
}
