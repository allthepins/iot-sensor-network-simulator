// Package nats provides NATS connection management and JetStream stream configuration.
package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	natsio "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	// DefaultStreamName is the name of the JetStream stream for sensor data.
	DefaultStreamName = "IOT_SENSORS"
	// DefaultSubjectPrefix is the prefix for all sensor subjects.
	DefaultSubjectPrefix = "iot.sensors"
)

// Client manages the NATS connection and JetStream operations.
type Client struct {
	conn   *natsio.Conn
	js     jetstream.JetStream
	logger *slog.Logger
}

// Config holds configuration for the NATS client.
type Config struct {
	URL            string
	StreamName     string
	SubjectPrefix  string
	MaxAge         time.Duration
	MaxMessages    int64
	ConnectTimeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		URL:            natsio.DefaultURL,
		StreamName:     DefaultStreamName,
		SubjectPrefix:  DefaultSubjectPrefix,
		MaxAge:         24 * time.Hour,
		MaxMessages:    10_000_000,
		ConnectTimeout: 10 * time.Second,
	}
}

// NewClient creates a new NATS client, establishes a connection,
// and configures the JetStream stream.
func NewClient(cfg Config, logger *slog.Logger) (*Client, error) {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("component", "nats_client")

	opts := []natsio.Option{
		natsio.Name("iot-simulator"),
		natsio.Timeout(cfg.ConnectTimeout),
		natsio.MaxReconnects(-1), // Infinite reconnect attempts
		natsio.DisconnectErrHandler(func(nc *natsio.Conn, err error) {
			if err != nil {
				logger.Warn("NATS disconnected", "error", err)
			}
		}),
		natsio.ReconnectHandler(func(nc *natsio.Conn) {
			logger.Info("NATS reconnected", "url", nc.ConnectedUrl())
		}),
	}

	conn, err := natsio.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logger.Info("Connected to NATS", "url", cfg.URL)

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	client := &Client{
		conn:   conn,
		js:     js,
		logger: logger,
	}

	// TODO: create or update stream
	if err := client.configureStream(cfg); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to configure stream: %w", err)
	}

	return client, nil
}

// configureStream creates or updates the JetStream stream config.
func (c *Client) configureStream(cfg Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	streamConfig := jetstream.StreamConfig{
		Name:        cfg.StreamName,
		Description: "IoT sensor data stream with 24-hour retention",
		Subjects:    []string{fmt.Sprintf("%s.>", cfg.SubjectPrefix)},
		MaxAge:      cfg.MaxAge,
		MaxMsgs:     cfg.MaxMessages,
		Discard:     jetstream.DiscardOld,
	}

	// Try to create stream
	stream, err := c.js.CreateStream(ctx, streamConfig)
	if err != nil {
		// If stream already exists, update it
		stream, err = c.js.UpdateStream(ctx, streamConfig)
		if err != nil {
			return fmt.Errorf("failed to create or update stream: %w", err)
		}
		c.logger.Info("Updated JetStream stream", "stream", cfg.StreamName)
	} else {
		c.logger.Info("Created JetStream stream", "stream", cfg.StreamName)
	}

	// Log stream state info
	// (Research says this is good practice)
	info, err := stream.Info(ctx)
	if err != nil {
		c.logger.Warn("Failed to get stream info", "error", err)
	} else {
		c.logger.Info("Stream configured",
			"messages", info.State.Msgs,
			"bytes", info.State.Bytes,
			"first_seq", info.State.FirstSeq,
			"last_seq", info.State.LastSeq,
		)
	}

	return nil
}

// Publish publishes a message to the specified subject.
func (c *Client) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := c.js.Publish(ctx, subject, data)
	return err
}

// PublishJson publishes a JSON-encoded message to the specified subject.
func (c *Client) PublishJson(ctx context.Context, subject string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return c.Publish(ctx, subject, data)
}

// Close gracefully closes the NATS connection.
func (c *Client) Close() error {
	if c.conn != nil {
		c.logger.Info("Closing NATS connection")
		c.conn.Close()
	}
	return nil
}

// IsConnected return true if the NATS connection is established.
func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

// Stats returns current connection statistics.
func (c *Client) Stats() natsio.Statistics {
	if c.conn == nil {
		return natsio.Statistics{}
	}
	return c.conn.Stats()
}

// JetStream returns the underlying JetStream context for advanced operations.
func (c *Client) JetStream() jetstream.JetStream {
	return c.js
}
