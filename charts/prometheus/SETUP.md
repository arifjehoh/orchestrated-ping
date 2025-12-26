# Prometheus Setup Guide

## Overview

Prometheus is an open-source monitoring and alerting toolkit designed for reliability and scalability in cloud-native environments. This guide walks through integrating Prometheus with the orchestrated-ping application.

## What is Prometheus?

**Key Features:**
- **Time-series database** - Stores metrics with timestamps
- **Pull-based model** - Scrapes metrics from targets
- **PromQL** - Powerful query language for metrics
- **Service discovery** - Automatically finds targets in Kubernetes
- **Alerting** - Alert based on metric thresholds

**How it works with go-ping:**
1. Prometheus scrapes `/metrics` endpoint from pods
2. Stores metrics in time-series database
3. Provides query interface for analysis
4. Can trigger alerts based on thresholds

## Architecture

```
go-ping Pods
    ↓ (expose /metrics)
Prometheus Server
    ↓ (scrape every 30s)
Time-Series Database
    ↓ (query with PromQL)
Grafana / Alerts
```

---

## Implementation Steps

### Step 1: Add Metrics Endpoint to go-ping

#### 1.1 Install Prometheus Client Library

```bash
cd api
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

#### 1.2 Create Metrics Package

Create `api/internal/metrics/metrics.go`:

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP request duration in seconds
    HttpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "Duration of HTTP requests in seconds",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "endpoint", "status"})

    // HTTP requests total
    HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"method", "endpoint", "status"})

    // Application uptime
    AppUptime = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "app_uptime_seconds",
        Help: "Application uptime in seconds",
    })
)
```

#### 1.3 Add Metrics Middleware

Update `api/internal/middleware/logger.go` or create new metrics middleware:

```go
package middleware

import (
    "net/http"
    "strconv"
    "time"

    "github.com/arifjehoh/orchestrated-ping/internal/metrics"
    "github.com/go-chi/chi/v5"
)

func Metrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap ResponseWriter to capture status code
        ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next.ServeHTTP(ww, r)
        
        duration := time.Since(start).Seconds()
        endpoint := chi.RouteContext(r.Context()).RoutePattern()
        status := strconv.Itoa(ww.statusCode)
        
        metrics.HttpDuration.WithLabelValues(r.Method, endpoint, status).Observe(duration)
        metrics.HttpRequestsTotal.WithLabelValues(r.Method, endpoint, status).Inc()
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

#### 1.4 Expose /metrics Endpoint

Update `api/internal/server/server.go`:

```go
package server

import (
    // ... existing imports
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupRouter(logger *slog.Logger, handler *handlers.Handler) *chi.Mux {
    r := chi.NewRouter()

    r.Use(chimiddleware.RequestID)
    r.Use(chimiddleware.RealIP)
    r.Use(middleware.Logger(logger))
    r.Use(middleware.Metrics)  // Add metrics middleware
    r.Use(chimiddleware.Recoverer)
    r.Use(chimiddleware.Timeout(60 * time.Second))

    r.Get("/ping", handler.Ping)
    r.Get("/health", handler.Health)
    r.Get("/ready", handler.Ready)
    r.Handle("/metrics", promhttp.Handler())  // Prometheus metrics endpoint

    return r
}
```

#### 1.5 Update Uptime Metric

In `api/main.go`, add a goroutine to update uptime:

```go
// After startTime := time.Now()
go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    for range ticker.C {
        metrics.AppUptime.Set(time.Since(startTime).Seconds())
    }
}()
```

#### 1.6 Test Metrics Locally

```bash
# Start the application
cd api
go run main.go

# Test metrics endpoint
curl http://localhost:8080/metrics

# Generate some requests
for i in {1..10}; do curl http://localhost:8080/ping; done

# Check metrics again
curl http://localhost:8080/metrics | grep http_requests_total
```

**Expected output:**
```
# HELP ping_requests_total Total number of ping requests
# TYPE ping_requests_total counter
http_requests_total{endpoint="/ping",method="GET",status="200"} 10
```

### Step 2: Install Prometheus using Helm

#### 2.1 Add Prometheus Helm Repository

```bash
# Add repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Search for available charts
helm search repo prometheus
```

#### 2.2 Review Default Values

```bash
# Download default values for reference
helm show values prometheus-community/prometheus > prometheus-defaults.yaml
```

#### 2.3 Create Custom Values File

Create `charts/prometheus/values.yaml`:

```yaml
# Disable unused components to save resources
alertmanager:
  enabled: false

# Disable node-exporter for minimal setup (enable for production)
nodeExporter:
  enabled: false

# Disable kube-state-metrics for minimal setup (enable for production)
kubeStateMetrics:
  enabled: false

# Disable pushgateway
pushgateway:
  enabled: false

# Prometheus server configuration
server:
  # Resource limits
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 250m
      memory: 256Mi
  
  # Data retention
  retention: "7d"
  
  # Persistent storage (optional for development)
  persistentVolume:
    enabled: false
    # Enable for production:
    # enabled: true
    # size: 10Gi
    # storageClass: standard
  
  # Service configuration
  service:
    type: ClusterIP
    port: 80
  
  # Scrape configuration
  global:
    scrape_interval: 30s
    scrape_timeout: 10s
    evaluation_interval: 30s

# Additional scrape configurations
serverFiles:
  prometheus.yml:
    scrape_configs:
      # Scrape Prometheus itself
      - job_name: prometheus
        static_configs:
          - targets:
            - localhost:9090

      # Scrape go-ping pods via pod annotations
      - job_name: 'kubernetes-pods'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          # Only scrape pods with prometheus.io/scrape=true annotation
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          # Use the port from prometheus.io/port annotation
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_port]
            action: replace
            target_label: __address__
            regex: ([^:]+)(?::\d+)?;(\d+)
            replacement: $1:$2
          # Use the path from prometheus.io/path annotation
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          # Add pod labels as metric labels
          - action: labelmap
            regex: __meta_kubernetes_pod_label_(.+)
          # Add namespace as label
          - source_labels: [__meta_kubernetes_namespace]
            action: replace
            target_label: kubernetes_namespace
          # Add pod name as label
          - source_labels: [__meta_kubernetes_pod_name]
            action: replace
            target_label: kubernetes_pod_name
```

#### 2.4 Deploy Prometheus

```bash
# Install Prometheus
helm install prometheus prometheus-community/prometheus \
  -f charts/prometheus/values.yaml \
  --namespace default

# Watch deployment
kubectl get pods -l app.kubernetes.io/name=prometheus -w

# Verify installation
helm status prometheus
```

#### 2.5 Access Prometheus UI

```bash
# Port-forward to access UI
kubectl port-forward svc/prometheus-server 9090:80 &

# Open in browser
open http://localhost:9090
```

### Step 3: Configure Service Discovery

#### Option A: Pod Annotations (Recommended for Simplicity)

Update `charts/go-ping/values.yaml`:

```yaml
# Pod annotations
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

Upgrade go-ping:
```bash
helm upgrade go-ping ./charts/go-ping
```

#### Option B: ServiceMonitor (Requires Prometheus Operator)

Create `charts/go-ping/templates/servicemonitor.yaml`:

```yaml
{{- if .Values.metrics.serviceMonitor.enabled }}
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: {{ include "go-ping.fullname" . }}
  labels:
    {{- include "go-ping.labels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "go-ping.selectorLabels" . | nindent 6 }}
  endpoints:
  - port: http
    path: /metrics
    interval: {{ .Values.metrics.serviceMonitor.interval }}
{{- end }}
```

Add to `charts/go-ping/values.yaml`:

```yaml
metrics:
  serviceMonitor:
    enabled: false
    interval: 30s
```

### Step 4: Verify Prometheus Scraping

#### 4.1 Check Targets

```bash
# Access Prometheus UI
kubectl port-forward svc/prometheus-server 9090:80

# Navigate to: http://localhost:9090/targets
# Look for "kubernetes-pods" job
# Verify go-ping pods are listed and UP
```

#### 4.2 Run Test Queries

Open Prometheus UI and run these queries:

```promql
# Check if go-ping is being scraped
up{kubernetes_pod_name=~"go-ping.*"}

# Request rate (last 5 minutes)
rate(http_requests_total[5m])

# Total ping requests
http_requests_total

# Request duration p95
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error rate
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
```

#### 4.3 Generate Load and Observe

```bash
# Generate requests
for i in {1..1000}; do curl -s http://localhost:8080/ping > /dev/null; done

# Query in Prometheus
rate(http_requests_total{endpoint="/ping"}[1m])
```

---

## NGINX Metrics (Optional)

To scrape metrics from nginx-proxy, enable stub_status:

Update `charts/nginx-proxy/templates/configmap.yaml`:

```nginx
# Add to server block
location /nginx-health {
    stub_status on;
    access_log off;
}
```

Add annotations to `charts/nginx-proxy/values.yaml`:

```yaml
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "80"
  prometheus.io/path: "/nginx-health"
```

**Note:** For better NGINX metrics, consider using [nginx-prometheus-exporter](https://github.com/nginxinc/nginx-prometheus-exporter).

---

## Alerting Rules

Create `charts/prometheus/alert-rules.yaml`:

```yaml
serverFiles:
  alerting_rules.yml:
    groups:
    - name: go-ping-alerts
      interval: 30s
      rules:
      - alert: HighErrorRate
        expr: |
          (
            sum(rate(http_requests_total{status=~"5.."}[5m]))
            /
            sum(rate(http_requests_total[5m]))
          ) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate on go-ping"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: HighResponseTime
        expr: |
          histogram_quantile(0.95, 
            rate(http_request_duration_seconds_bucket[5m])
          ) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time on go-ping"
          description: "p95 response time is {{ $value }}s"

      - alert: PodDown
        expr: up{job="kubernetes-pods", kubernetes_pod_name=~"go-ping.*"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "go-ping pod is down"
          description: "Pod {{ $labels.kubernetes_pod_name }} is down"
```

Apply with:
```bash
helm upgrade prometheus prometheus-community/prometheus \
  -f charts/prometheus/values.yaml \
  -f charts/prometheus/alert-rules.yaml
```

---

## Troubleshooting

### Targets Not Appearing

```bash
# Check Prometheus logs
kubectl logs -l app.kubernetes.io/name=prometheus

# Verify pod annotations
kubectl describe pod -l app.kubernetes.io/name=go-ping | grep Annotations

# Check RBAC permissions
kubectl get clusterrole prometheus-server -o yaml
```

### Metrics Not Scraped

```bash
# Test metrics endpoint from within cluster
kubectl run curl --image=curlimages/curl -it --rm -- \
  curl http://go-ping:8080/metrics

# Check Prometheus config
kubectl get configmap prometheus-server -o yaml
```

### High Cardinality Issues

```bash
# Check metric cardinality
# In Prometheus UI, query:
count({__name__=~".+"}) by (__name__)

# Avoid high-cardinality labels (user IDs, timestamps, etc.)
```

---

## Learning Resources

### Essential Reading
- [Prometheus Overview](https://prometheus.io/docs/introduction/overview/)
- [Metric Types](https://prometheus.io/docs/concepts/metric_types/)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Go Application Metrics](https://prometheus.io/docs/guides/go-application/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)

### Useful PromQL Queries
- [PromQL Cheat Sheet](https://promlabs.com/promql-cheat-sheet/)
- [Query Examples](https://prometheus.io/docs/prometheus/latest/querying/examples/)

---

## Future: OpenTelemetry Migration

Once you're comfortable with Prometheus, consider migrating to **OpenTelemetry** for modern observability:

### Why OpenTelemetry?

- **Vendor-neutral**: Not locked to Prometheus
- **Distributed tracing**: Track requests across services  
- **Unified observability**: Metrics + traces + logs in one SDK
- **Auto-instrumentation**: Less manual code
- **Industry standard**: CNCF graduated project

### Migration Approach

```go
// Instead of Prometheus client
import "github.com/prometheus/client_golang/prometheus"

// Use OpenTelemetry SDK
import "go.opentelemetry.io/otel/metric"
```

OpenTelemetry can export to Prometheus (backward compatible), plus add distributed tracing with Jaeger/Tempo.

**Architecture**:
```
go-ping (OTel SDK)
    ↓
OTel Collector
    ├─→ Prometheus (metrics)
    └─→ Jaeger (traces)
    ↓
Grafana (unified view)
```

This allows you to keep your existing Prometheus/Grafana setup while gaining advanced capabilities.

### Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [OTel Collector](https://opentelemetry.io/docs/collector/)
- [Prometheus + OTel](https://prometheus.io/docs/prometheus/latest/feature_flags/#otlp-receiver)

## Next Steps

1. ✅ Complete Prometheus setup
2. ⏭️ Set up Grafana for visualization
3. ⏭️ Create custom dashboards
4. ⏭️ Configure alerting with Alertmanager (optional)
5. ⏭️ Add persistence for production use
6. ⏭️ (Advanced) Explore OpenTelemetry migration

See [Grafana Setup Guide](../grafana/SETUP.md) for the next phase.
