// aggregator_test contains tests for the aggregator package.
package aggregator_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/aggregator"
	"github.com/allthepins/iot-sensor-network-simulator/internal/model"
)

// newTestLogger returns a slog.Logger to facilitate testing function log text.
func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	handler := slog.NewTextHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(handler)
}

// TestNewAggregator verifies that the New function correctly initializes an Aggregator.
func TestNewAggregator(t *testing.T) {
	t.Parallel()
	dataCh := make(chan model.SensorData)
	agg := aggregator.New(dataCh, nil, nil)

	if agg == nil {
		t.Fatal("New returned nil")
	}
	if agg.DataCh == nil {
		t.Error("DataCh was not initialized")
	}
}

// TestAggregator_Run_ProcesssesData verifies that the aggregator receives and logs data.
func TestAggregator_Run_ProcessesData(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	dataCh := make(chan model.SensorData, 1) // Buffer channel to prevent blocking
	agg := aggregator.New(dataCh, nil, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		agg.Run(ctx)
	}()

	// Send data to the data channel
	testData := model.SensorData{ID: 1, Value: 0.99}
	dataCh <- testData

	// Give the aggregator enough to process, so that the summary is logged.
	time.Sleep(6 * time.Second)

	if !strings.Contains(buf.String(), "count=1") {
		t.Errorf("expected log to contain summary of processed data, but it didn't. Log %s", buf.String())
	}

	cancel()
	wg.Wait()
}

// TestAggregator_Run_StopsOnContextCancel verified the aggregator stops when the context is canceled.
// TODO Confirm if receiving on `runFinished` properly confirms Run's graceful exit on context cancellation.
func TestAggregator_Run_StopsOnContextCancel(t *testing.T) {
	t.Parallel()
	dataCh := make(chan model.SensorData)
	agg := aggregator.New(dataCh, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())

	runFinished := make(chan struct{})
	go func() {
		agg.Run(ctx)
		close(runFinished)
	}()

	cancel()

	select {
	case <-runFinished:
		// Expected behavior: Run exited gracefully.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("aggregator did not stop after context cancellation")
	}
}

// TestAggregator_Run_StopsOnChannelClose verifies the aggregator stops when the data channel is closed.
func TestAggregator_Run_StopsOnChannelClose(t *testing.T) {
	t.Parallel()
	dataCh := make(chan model.SensorData)
	agg := aggregator.New(dataCh, nil, nil)
	ctx := context.Background()

	runFinished := make(chan struct{})
	go func() {
		agg.Run(ctx)
		close(runFinished)
	}()

	close(dataCh)

	select {
	case <-runFinished:
		// Expected behavior: Run exited gracefully, test passed.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("aggregator did not stop after channel was closed")
	}
}
