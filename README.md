# **IoT Sensor Network Simulator**

Simulate a network of IoT sensors (built in Go).

It concurrently runs thousands of virtual sensors that send data to a central aggregator.

## Features

- **High Concurrency:** Simulates thousands of goroutines concurrently.

- **Graceful Shutdown:** Stop the simulation gracefully at any time with `ctrl+c`.

- **Panic Recovery:** Sensors will restart automatically if they encounter a panic during operation.

- **Real-time monitoring:** Metrics can be monitored in real-time via Prometheus.

- **Pre-configured Dashboard:** Includes a Grafana dashboard for visualizing key metrics.

- **Built-in profiling:** Uses Go's built-in `pprof` tools to inspect performance in real time.

## Directory Structure
```
├── cmd/simulator/main.go   # Main application entry point.
├── internal/               # Private application packages.
│   ├── aggregator/         # Consumes and processes sensor data.
│   ├── metrics/            # Prometheus metric definitions.
│   ├── model/              # Shared data structures (e.g. SensorData).
│   ├── sensor/             # Simulates a single IoT sensor.
│   └── server/             # HTTP server for the metrics and pprof endpoints.
├── grafana/                # Grafana configuration.
├── go.mod                  # Go module definitions.
├── Dockerfile              # Build instructions for the iot-simulator Docker container image.
├── compose.yaml            # Docker Compose (V2) config.
├── prometheus.yml          # Prometheus scrape target configuration.
└── README.md               # You are here :)
```

## Getting Started

### Prerequisites

- Docker 28 or later.

### Usage

#### Running with Docker: 
The simulation runs for a pre-configured duration (default: 10 minutes) or until you stop it manually with `ctrl+c`.
1. **Build and run the simulator:**
```shell
docker compose up --build
```

2. **Observe the output:**

   You will see logs from the aggregator and sensors directly in your terminal.

3. **Stopping the simulation:**

   Press `ctrl+c` at anytime to initiate a graceful shutdown. The application will stop all sensors, process any remaining data, and then exit.

   To stop and remove the containers and their volumes, run:
   ```shell
   docker compose down -v
   ```

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

### Visualizing with Grafana

The stack includes a Grafana instance with a pre-built dashboard which provides an overview of the simulator's performance.

1. Access the Grafana UI at http://localhost:3000
2. Login using the credentials `admin`/`admin`.
3. Navigate to "Dashboards" to find the "IoT Simulator Dashboard".

The dashboard in addition to providing real-time insights also fires an alert if any sensor restarts within a 5-minute window.

### Viewing Metrics directly in Prometheus

Prometheus automatically scrapes metrics exposed by the simulator at http://simulator:2112/metrics (this endpoint has only been exposed internally within Docker).

To explore the collected metrics:

1. Access the Prometheus UI at http://localhost:9090
2. Use the "Expression" input field to enter PromQL queries.
3. Use the "Graph" tab to visualize query results. 

#### Example PromQL queries

*General Overview*

| Query                                                        | Description                                   |
| ------------------------------------------------------------ | --------------------------------------------- |
| `iot_simulator_active_sensors`                               | Current number of active sensor goroutines    |
| `iot_simulator_aggregator_messages_received_total`           | Total messages received by the aggregator     |
| `rate(iot_simulator_aggregator_messages_received_total[1m])` | Message ingestion rate over the last 1 minute |

*Per-Sensor Metrics*

| Query                                                                                                      | Description                                          |
| ---------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| `iot_simulator_sensor_messages_sent_total`                                                                 | Total messages sent per sensor                       |
| `rate(iot_simulator_sensor_messages_sent_total[5m])`                                                       | Message send rate per sensor over the last 5 minutes |
| `iot_simulator_sensor_generated_values_bucket`                                                             | Distribution of generated values (histogram)         |
| `histogram_quantile(0.95, sum(rate(iot_simulator_sensor_generated_values_bucket[1m])) by (le, sensor_id))` | 95th percentile of generated values per sensor       |

*Failures/Restarts*

| Query                                                | Description                                     |
| ---------------------------------------------------- | ----------------------------------------------- |
| `iot_simulator_sensor_restarts_total`                | Number of restarts per sensor due to panics     |
| `increase(iot_simulator_sensor_restarts_total[10m])` | Restart count per sensor in the last 10 minutes |

### Profiling with `pprof`

The simulator includes built-in support for `pprof` and exposes a profiling HTTP server on http://localhost:6060/debug/pprof.

This allows analysis of CPU usage, memory allocations, goroutine activity, blocking calls etc. during runtime.

One can also analyze data via the command line using `go tool pprof`. Some examples include:

Capture a CPU profile:
```shell
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

Inspect in interactive mode:
```shell
(pprof) top        # View top functions by CPU usage
(pprof) list func  # View annotated source of specific function
(pprof) web        # Opens visualization in your browser (requires Graphviz)
```

---

## **Roadmap**

### **Core Simulator**

- [x] Basic sensor goroutines (simulate value, send message)
- [x] Graceful shutdown support
- [x] Panic recovery & auto-restart per sensor

### **Observability**

- [x] Prometheus metrics: messages, restarts, values
- [x] `/metrics` HTTP endpoint
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
