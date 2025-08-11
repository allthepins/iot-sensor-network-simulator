# **IoT Sensor Network Simulator**

Simulate a network of IoT sensors (built in Go).

It concurrently runs thousands of virtual sensors that send data to a central aggregator.

## Features

- **High Concurrency:** Simulates thousands of goroutines concurrently.

- **Graceful Shutdown:** Stop the simulation gracefully at any time with `ctrl+c`.

- **Panic Recovery:** Sensors will restart automatically if they encounter a panic during operation.

## Directory Structure
```
├── cmd/simulator/main.go   # Main application entry point.
├── internal/               # Private application packages.
│   ├── aggregator/         # Consumes and processes sensor data.
│   ├── model/              # Shared data structures (e.g. SensorData).
│   └── sensor/             # Simulates a single IoT sensor.
├── go.mod                  # Go module definitions.
├── Dockerfile              # Build instructions for the iot-simulator Docker container image.
├── compose.yaml            # Docker Compose (V2) config.
└── README.md               # You are here :)
```

## Getting Started

### Prerequisites

- Docker 28 or later.

### Usage

#### Running with docker: 
The simulation runs for a configured diratio (default: 10 seconds) or until you stop it manually with `ctrl+c`.
1. **Build and run the simulator:**
```shell
docker compose up --build
```

2. **Observe the output:**

   You will see logs from the aggregator and sensors directly in your terminal.

3. **Stopping the simulation:**

   Press `ctrl+c` at anytime to initiate a graceful shutdown. The application will stop all sensors, process any remaining data, and then exit.

#### Running natively (local development):
This requires Go 1.18 or later.

The application can be run direcly using the `go run` command. It runs for the configured duration or until you stop it manually with `ctrl+c`.
```shell
go run ./cmd/simulator
```

### Running Tests

To run the unit tests for all packages, execute the following command form the root directory:

```shell
go test -v ./...
```

(This will discover and run all files ending in _test.go)

---

## **Roadmap**

### **Core Simulator**

- [x] Basic sensor goroutines (simulate value, send message)
- [x] Graceful shutdown support
- [x] Panic recovery & auto-restart per sensor

### **Observability**

- [ ] Prometheus metrics: messages, restarts, values
- [ ] `/metrics` HTTP endpoint
- [ ] Grafana dashboard
- [ ] Alert: high restart count
- [ ] Time series: sensor value by ID
- [ ] Table: per-sensor message counts

### **Publish/Subscribe**

- [ ] NATS publisher integration
- [ ] Basic aggregator subscriber

### **Nice to haves**

- [ ] Sensor types: temperature, humidity, battery, etc.
- [ ] Simulated failures (such as random drops, latency)
- [ ] Metadata injection (such as location)
- [ ] Distributed sensor runner (deploy across multiple machines)
- [ ] Historical replay mode (simulate past data)
- [ ] API to control sensors live

### DevOps/Scaling

- [ ] Load testing
- [ ] TBD
