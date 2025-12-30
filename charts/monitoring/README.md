# Monitoring Integration Overview

This directory contains setup guides and configurations for integrating Prometheus and Grafana with the orchestrated-ping application.

## Structure

```
charts/
├── prometheus/
│   ├── SETUP.md              # Prometheus installation and configuration
│   ├── values.yaml           # Helm values for Prometheus
│   └── alert-rules.yaml      # Alerting rules (optional)
├── grafana/
│   ├── SETUP.md              # Grafana installation and configuration
│   ├── values.yaml           # Helm values for Grafana
│   └── dashboards/           # Custom dashboard JSON files
│       ├── go-ping-dashboard.json
│       ├── nginx-dashboard.json
│       └── cluster-dashboard.json
└── monitoring/
    └── README.md             # This file
```

## Implementation Phases

### Phase 1: Prometheus Setup
**Goal**: Collect metrics from go-ping application

**Steps**:
1. Add `/metrics` endpoint to go-ping
2. Install Prometheus via Helm
3. Configure service discovery
4. Verify scraping

**Time**: ~4-6 hours

📖 See [Prometheus Setup Guide](../prometheus/SETUP.md)

### Phase 2: Grafana Setup
**Goal**: Visualize metrics in dashboards

**Steps**:
1. Install Grafana via Helm
2. Configure Prometheus data source
3. Create custom dashboards
4. Import community dashboards

**Time**: ~3-4 hours

📖 See [Grafana Setup Guide](../grafana/SETUP.md)

### Phase 3: Advanced Features (Optional)
**Goal**: Production-ready monitoring

**Features**:
- Alerting and notifications
- Data persistence
- External access (Ingress)
- Additional exporters

**Time**: ~2-3 hours

### Phase 4: OpenTelemetry Migration (Advanced)
**Goal**: Modern observability with vendor-neutral instrumentation

**Features**:
- Replace Prometheus client with OpenTelemetry SDK
- Deploy OpenTelemetry Collector
- Enable distributed tracing with Jaeger/Tempo
- Unified metrics, traces, and logs
- Maintain Prometheus/Grafana compatibility

**Benefits**:
- Vendor-neutral telemetry data
- Distributed tracing across services
- Auto-instrumentation capabilities
- Industry-standard CNCF solution
- Correlate metrics, traces, and logs

**Time**: ~6-8 hours

📖 See [OpenTelemetry Migration Guide](../opentelemetry/MIGRATION.md) (future)

## Quick Start

### Minimal Setup (Development)

```bash
# 1. Install Prometheus
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/prometheus \
  -f charts/prometheus/values.yaml

# 2. Install Grafana
helm repo add grafana https://grafana.github.io/helm-charts
helm install grafana grafana/grafana \
  -f charts/grafana/values.yaml

# 3. Access UIs
kubectl port-forward svc/prometheus-server 9090:80 &
kubectl port-forward svc/grafana 3000:80 &

# Open browsers:
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
```

## Learning Path

### Before You Start

Read these topics to understand the fundamentals:

#### Prometheus
- [ ] [What is Prometheus?](https://prometheus.io/docs/introduction/overview/)
- [ ] [Metric Types](https://prometheus.io/docs/concepts/metric_types/) - Counter, Gauge, Histogram, Summary
- [ ] [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)

#### Grafana
- [ ] [Grafana Overview](https://grafana.com/docs/grafana/latest/introduction/)
- [ ] [Dashboard Basics](https://grafana.com/docs/grafana/latest/dashboards/)
- [ ] [Using Prometheus in Grafana](https://grafana.com/docs/grafana/latest/datasources/prometheus/)

#### Monitoring Philosophy
- [ ] [RED Method](https://grafana.com/blog/2018/08/02/the-red-method-how-to-instrument-your-services/) - Rate, Errors, Duration
- [ ] [USE Method](https://www.brendangregg.com/usemethod.html) - Utilization, Saturation, Errors
- [ ] [Four Golden Signals](https://sre.google/sre-book/monitoring-distributed-systems/)

### Hands-on Exercises

1. **Explore Prometheus UI**
   - Browse targets
   - Run basic queries
   - Understand time series

2. **Build a Simple Dashboard**
   - Create panels
   - Use different visualizations
   - Add variables

3. **Set Up an Alert**
   - Define alert rule
   - Configure notification channel
   - Test alert triggering

## Architecture Diagrams

### Current State (Without Monitoring)

```
nginx-proxy → go-ping Service → go-ping Pods
```

### Target State (With Monitoring)

```
┌─────────────────────────────────────────────┐
│                  Users                      │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│              nginx-proxy                    │
│         (with metrics endpoint)             │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│               go-ping Pods                  │
│         (expose /metrics endpoint)          │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│           Prometheus Server                 │
│    (scrapes metrics every 30s)              │
│    (stores in time-series DB)               │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│               Grafana                       │
│    (queries Prometheus with PromQL)         │
│    (displays in dashboards)                 │
└─────────────────────────────────────────────┘
                    ↑
┌─────────────────────────────────────────────┐
│          Operators/Developers               │
└─────────────────────────────────────────────┘
```

## Key Metrics to Monitor

### Application Metrics (RED Method)

- **Rate**: Requests per second
  ```promql
  rate(http_requests_total[5m])
  ```

- **Errors**: Error rate percentage
  ```promql
  sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
  ```

- **Duration**: Response time (p95, p99)
  ```promql
  histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
  ```

### Infrastructure Metrics (USE Method)

- **Utilization**: CPU/Memory usage
- **Saturation**: Queue depth, wait time
- **Errors**: Failed requests, pod restarts

## Testing Your Setup

### 1. Verify Metrics Endpoint

```bash
# Test locally
curl http://localhost:8080/metrics

# Test in cluster
kubectl run curl --image=curlimages/curl -it --rm -- \
  curl http://go-ping:8080/metrics
```

### 2. Generate Load

```bash
# Simple load generation
for i in {1..1000}; do 
  curl -s http://localhost:8080/ping > /dev/null
done

# With K6 (if installed)
k6 run k6-load-test.js
```

### 3. Query Metrics

```promql
# In Prometheus UI (http://localhost:9090)

# Total requests
http_requests_total

# Request rate
rate(http_requests_total[5m])

# Pod uptime
app_uptime_seconds
```

### 4. View in Grafana

- Navigate to http://localhost:3000
- Explore > Select Prometheus
- Run queries and visualize

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Metrics not appearing | Check pod annotations, verify scraping in Prometheus |
| Grafana can't connect to Prometheus | Verify service name and port |
| High cardinality warning | Reduce number of labels in metrics |
| Slow queries | Reduce time range, use recording rules |
| Dashboard not loading | Check JSON syntax, verify data source |

### Debug Commands

```bash
# Check Prometheus targets
kubectl port-forward svc/prometheus-server 9090:80
# Navigate to http://localhost:9090/targets

# View Prometheus config
kubectl get configmap prometheus-server -o yaml

# Check Grafana logs
kubectl logs -l app.kubernetes.io/name=grafana

# Test data source from Grafana pod
kubectl exec -it <grafana-pod> -- wget -O- http://prometheus-server:80/-/healthy
```

## Production Considerations

### Before Going to Production

- [ ] Enable persistence for both Prometheus and Grafana
- [ ] Configure proper retention policies
- [ ] Set up authentication and authorization
- [ ] Configure HTTPS/TLS
- [ ] Set up alerting with proper channels
- [ ] Add resource limits based on load testing
- [ ] Configure backup and restore procedures
- [ ] Document runbooks for alerts

### Security

- [ ] Change default passwords
- [ ] Use secrets for sensitive data
- [ ] Enable RBAC
- [ ] Restrict network access
- [ ] Enable audit logging

### High Availability

- [ ] Run multiple Prometheus replicas
- [ ] Configure persistent storage
- [ ] Set up Alertmanager cluster
- [ ] Consider Thanos for long-term storage

## Additional Resources

### Official Documentation
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [prometheus-community Helm Charts](https://github.com/prometheus-community/helm-charts)
- [Grafana Helm Chart](https://github.com/grafana/helm-charts)

### Community Resources
- [Awesome Prometheus](https://github.com/roaldnefs/awesome-prometheus)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- [PromQL for Humans](https://timber.io/blog/promql-for-humans/)

### Books
- "Prometheus: Up & Running" by Brian Brazil
- "Monitoring with Prometheus" by James Turnbull

## Migration Path to OpenTelemetry

Once you're comfortable with Prometheus and Grafana, consider migrating to OpenTelemetry for modern observability:

### Why OpenTelemetry?

**Current approach (Prometheus native)**:
```
go-ping → /metrics endpoint → Prometheus → Grafana
```

**OpenTelemetry approach**:
```
go-ping (OTel SDK)
    ↓
OpenTelemetry Collector
    ├─→ Prometheus (metrics)
    ├─→ Jaeger/Tempo (traces) 
    └─→ Loki (logs)
    ↓
Grafana (unified observability)
```

### Advantages

- **Vendor-neutral**: Not locked into Prometheus
- **Distributed tracing**: Track requests across microservices
- **Unified SDK**: Metrics, traces, logs in one library
- **Auto-instrumentation**: Automatic HTTP, gRPC, database tracking
- **Industry standard**: CNCF graduated project
- **Future-proof**: Wide adoption and active development

### When to Migrate

✅ **Migrate when**:
- Building multiple microservices
- Need distributed tracing
- Want flexibility to switch backends
- Ready for production observability

⏸️ **Stay with Prometheus if**:
- Single service/simple architecture
- Only need basic metrics
- Learning fundamentals
- Prototyping quickly

### Migration Strategy

1. **Phase 1**: Keep Prometheus, add OTel SDK alongside
2. **Phase 2**: Deploy OTel Collector, export to Prometheus
3. **Phase 3**: Add tracing backend (Jaeger/Tempo)
4. **Phase 4**: Remove direct Prometheus client (optional)

This allows gradual migration with zero downtime.

### Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OTel Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [OTel Collector](https://opentelemetry.io/docs/collector/)
- [Prometheus OTLP Receiver](https://prometheus.io/docs/prometheus/latest/feature_flags/#otlp-receiver)

## Next Steps

1. Complete [Prometheus Setup](../prometheus/SETUP.md)
2. Complete [Grafana Setup](../grafana/SETUP.md)
3. Create custom dashboards
4. Configure alerting
5. Document your setup
6. (Optional) Explore OpenTelemetry migration

## Success Criteria

✅ Prometheus successfully scrapes go-ping metrics  
✅ Grafana displays real-time metrics  
✅ Dashboards are created and working  
✅ Alerts can be configured (optional)  
✅ Documentation is complete  

Happy monitoring! 📊
