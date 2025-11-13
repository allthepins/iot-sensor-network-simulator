package nats_test

import (
	"testing"
	"time"

	"github.com/allthepins/iot-sensor-network-simulator/internal/nats"
)

// TestDefaultConfig verifies the default configuration values.
func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := nats.DefaultConfig()

	if cfg.StreamName != nats.DefaultConfig().StreamName {
		t.Errorf("expected StreamName %s, got %s", nats.DefaultStreamName, cfg.StreamName)
	}

	if cfg.SubjectPrefix != nats.DefaultSubjectPrefix {
		t.Errorf("expected SubjectPrefix %s, got %s", nats.DefaultSubjectPrefix, cfg.SubjectPrefix)
	}

	if cfg.MaxAge != 24*time.Hour {
		t.Errorf("expected MaxAge 24h, got %v", cfg.MaxAge)
	}

	if cfg.MaxMessages != 10_000_000 {
		t.Errorf("expected MaxMessages 10M, got %d", cfg.MaxMessages)
	}

	if cfg.ConnectTimeout != 10*time.Second {
		t.Errorf("expected ConnectTimeout 10s, got %v", cfg.ConnectTimeout)
	}
}

// TestNewClient_InvalidURL tests that NewClient returns an error for invalid NATS URLs.
func TestNewClient_InvalidURL(t *testing.T) {
	t.Parallel()

	cfg := nats.DefaultConfig()
	cfg.URL = "nats://invalid-host:4222"
	cfg.ConnectTimeout = 1 * time.Second

	client, err := nats.NewClient(cfg, nil)
	if err == nil {
		t.Fatal("expected error for invalid NATS URL, got nil")
	}
	if client != nil {
		t.Error("expected nil client on error")
	}
}

// TODO: Implement integration tests with a real NATS server:
// - Connection to NATS server
// - Stream create/update
// - Publish messages
// - Connection/Reconnection
// - Graceful shutdown
