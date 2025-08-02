# **IoT Sensor Network Simulator**

Simulate a network of IoT sensors (built in Go).

It simulates thousands of sensor nodes publishing messages over NATS.

---

## **Roadmap**

### **Core Simulator**

- [ ] Basic sensor goroutines (simulate value, send message)
- [ ] Graceful shutdown support
- [ ] Panic recovery & auto-restart per sensor

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
