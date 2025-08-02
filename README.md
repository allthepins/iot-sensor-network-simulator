# IoT Sensor Network Simulator

Simulate a network of IoT sensors (built in Go).

It simulates thousands of sensor nodes publishing messages over NATS.

---

## Roadmap

### Core Simulator

- [ ] Basic sensor goroutines (simulate value, send message)
- [ ] NATS publisher integration
- [ ] Graceful shutdown support
- [ ] Panic recovery & auto-restart per sensor
- [ ] Basic aggregator subscriber

### Observability

- [ ] Prometheus metrics: messages, restarts, values
- [ ] TBD

### Nice to haves

- [ ] Sensor types: temperature, humidity, battery, etc.
- [ ] Simulated failures (such as random drops, latency)
- [ ] Metadata injection (such as location)
- [ ] Distributed sensor runner (deploy across multiple machines)
- [ ] Historical replay mode (simulate past data)
- [ ] API to control sensors live

### DevOps/Scaling

- [ ] Load testing
- [ ] TBD
