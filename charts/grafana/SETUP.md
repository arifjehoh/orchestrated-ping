# Grafana Setup Guide

## Overview

Grafana is an open-source analytics and monitoring platform that visualizes metrics from Prometheus and other data sources. This guide walks through setting up Grafana to visualize orchestrated-ping metrics.

## What is Grafana?

**Key Features:**
- **Dashboard creation** - Visual representation of metrics
- **Multiple data sources** - Prometheus, InfluxDB, Elasticsearch, etc.
- **Alerting** - Visual alerts and notifications
- **Templating** - Dynamic, reusable dashboards
- **Sharing** - Export and share dashboards as JSON

**Integration with Prometheus:**
1. Grafana connects to Prometheus as data source
2. Queries metrics using PromQL
3. Displays data in customizable dashboards
4. Provides alerting and notifications

## Architecture

```
Prometheus
    ↓ (data source)
Grafana
    ↓ (query PromQL)
Dashboards
    ↑
Users access via browser
```

---

## Implementation Steps

### Step 1: Install Grafana using Helm

#### 1.1 Add Grafana Helm Repository

```bash
# Add repository
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Search for chart
helm search repo grafana/grafana
```

#### 1.2 Review Default Values

```bash
# Download default values
helm show values grafana/grafana > grafana-defaults.yaml
```

#### 1.3 Create Custom Values File

Create `charts/grafana/values.yaml`:

```yaml
# Admin credentials
adminUser: admin
adminPassword: admin  # Change in production!

# Resources
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi

# Service configuration
service:
  type: ClusterIP
  port: 80

# Persistence (optional for development)
persistence:
  enabled: false
  # Enable for production:
  # enabled: true
  # size: 5Gi
  # storageClass: standard

# Data sources configuration
datasources:
  datasources.yaml:
    apiVersion: 1
    datasources:
    - name: Prometheus
      type: prometheus
      url: http://prometheus-server:80
      access: proxy
      isDefault: true
      jsonData:
        timeInterval: 30s

# Dashboard providers
dashboardProviders:
  dashboardproviders.yaml:
    apiVersion: 1
    providers:
    - name: 'default'
      orgId: 1
      folder: ''
      type: file
      disableDeletion: false
      editable: true
      options:
        path: /var/lib/grafana/dashboards/default

# Pre-load dashboards
dashboards:
  default:
    # We'll add custom dashboards here later
    # kubernetes-cluster:
    #   url: https://grafana.com/api/dashboards/7249/revisions/1/download

# Plugins to install
plugins: []
  # - grafana-piechart-panel
  # - grafana-clock-panel

# Environment variables
env: {}

# Ingress (optional)
ingress:
  enabled: false
  # Enable for external access:
  # enabled: true
  # hosts:
  #   - grafana.example.com
  # tls:
  #   - secretName: grafana-tls
  #     hosts:
  #       - grafana.example.com

# Security context
securityContext:
  runAsUser: 472
  runAsGroup: 472
  fsGroup: 472
```

#### 1.4 Deploy Grafana

```bash
# Install Grafana
helm install grafana grafana/grafana \
  -f charts/grafana/values.yaml \
  --namespace default

# Watch deployment
kubectl get pods -l app.kubernetes.io/name=grafana -w

# Verify installation
helm status grafana
```

#### 1.5 Access Grafana UI

```bash
# Get admin password (if not set in values)
kubectl get secret grafana -o jsonpath="{.data.admin-password}" | base64 --decode
echo

# Port-forward to access UI
kubectl port-forward svc/grafana 3000:80 &

# Open in browser
open http://localhost:3000

# Login with:
# Username: admin
# Password: admin (or from secret above)
```

### Step 2: Configure Prometheus Data Source

#### Option A: Automatic (via Helm values)

Already configured in `values.yaml` above! Verify in Grafana:

1. Go to Configuration > Data Sources
2. Click on "Prometheus"
3. Scroll down and click "Save & Test"
4. Should see "Data source is working"

#### Option B: Manual Configuration

If not using Helm values:

1. Login to Grafana (http://localhost:3000)
2. Go to Configuration > Data Sources
3. Click "Add data source"
4. Select "Prometheus"
5. Configure:
   - **Name**: Prometheus
   - **URL**: `http://prometheus-server:80`
   - **Access**: Server (proxy)
6. Click "Save & Test"

### Step 3: Create Custom Dashboards

#### Dashboard 1: go-ping Application Metrics

Create `charts/grafana/dashboards/go-ping-dashboard.json`:

**Manual creation steps:**

1. Click "+" > "Dashboard"
2. Click "Add new panel"
3. Add these panels:

**Panel 1: Request Rate**
- Query: `rate(http_requests_total{endpoint="/ping"}[5m])`
- Visualization: Graph/Time series
- Title: "Requests per Second"
- Unit: reqps (requests/sec)

**Panel 2: Request Duration (p95)**
- Query: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- Visualization: Graph/Time series
- Title: "Response Time p95"
- Unit: seconds

**Panel 3: Error Rate**
- Query: `sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))`
- Visualization: Stat
- Title: "Error Rate"
- Unit: percentunit (0-1)

**Panel 4: Total Requests**
- Query: `sum(http_requests_total)`
- Visualization: Stat
- Title: "Total Requests"

**Panel 5: Application Uptime**
- Query: `app_uptime_seconds`
- Visualization: Stat
- Title: "Uptime"
- Unit: seconds

**Panel 6: Pod Count**
- Query: `count(up{kubernetes_pod_name=~"go-ping.*"})`
- Visualization: Stat
- Title: "Running Pods"

4. Save dashboard as "go-ping Application Metrics"

#### Export Dashboard JSON

```bash
# After creating dashboard in UI:
# 1. Click "Dashboard settings" (gear icon)
# 2. Click "JSON Model"
# 3. Copy the JSON
# 4. Save to charts/grafana/dashboards/go-ping-dashboard.json
```

#### Load Dashboard via Helm

Update `charts/grafana/values.yaml`:

```yaml
dashboards:
  default:
    go-ping:
      file: dashboards/go-ping-dashboard.json
```

Upgrade Grafana:
```bash
helm upgrade grafana grafana/grafana -f charts/grafana/values.yaml
```

### Step 4: Import Community Dashboards

#### Import via UI

1. Click "+" > "Import"
2. Enter Dashboard ID
3. Click "Load"
4. Select Prometheus data source
5. Click "Import"

#### Recommended Dashboards

**Kubernetes Cluster Monitoring** (ID: 7249)
```bash
# Or import via Helm values.yaml:
dashboards:
  default:
    kubernetes-cluster:
      gnetId: 7249
      revision: 1
      datasource: Prometheus
```

**Node Exporter Full** (ID: 1860)
- Detailed node metrics (requires node-exporter)

**Go Processes** (ID: 6671)
- Go runtime metrics

**NGINX Ingress Controller** (ID: 9614)
- NGINX metrics (if using nginx-proxy)

### Step 5: Create Alert Rules in Grafana

#### Configure Alert Channel

1. Go to Alerting > Notification channels
2. Click "Add channel"
3. Configure (e.g., Email, Slack, Webhook)
4. Test and save

#### Add Alert to Panel

1. Edit any panel (e.g., "Error Rate")
2. Click "Alert" tab
3. Click "Create Alert"
4. Configure:
   - **Condition**: WHEN avg() OF query(A, 5m, now) IS ABOVE 0.05
   - **Frequency**: Evaluate every 1m
   - **For**: 5m
5. Add notification channel
6. Save

---

## Dashboard Templates

### go-ping Application Dashboard

Basic structure:

```json
{
  "dashboard": {
    "title": "go-ping Application Metrics",
    "panels": [
      {
        "title": "Requests per Second",
        "targets": [
          {
            "expr": "rate(http_requests_total{endpoint=\"/ping\"}[5m])"
          }
        ]
      },
      {
        "title": "Response Time p95",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ]
      }
    ]
  }
}
```

Full dashboard JSON will be created during implementation.

### Variables for Dynamic Dashboards

Add variables for reusability:

1. Dashboard settings > Variables
2. Add variable:
   - **Name**: namespace
   - **Type**: Query
   - **Query**: `label_values(up, kubernetes_namespace)`
3. Use in queries: `{kubernetes_namespace="$namespace"}`

---

## Advanced Configuration

### Enable Plugins

Update `values.yaml`:

```yaml
plugins:
  - grafana-piechart-panel
  - grafana-clock-panel
```

### Configure SMTP for Alerts

```yaml
env:
  GF_SMTP_ENABLED: true
  GF_SMTP_HOST: smtp.gmail.com:587
  GF_SMTP_USER: your-email@gmail.com
  GF_SMTP_PASSWORD: your-app-password
  GF_SMTP_FROM_ADDRESS: your-email@gmail.com
```

### Enable Anonymous Access (Read-only)

```yaml
grafana.ini:
  auth.anonymous:
    enabled: true
    org_role: Viewer
```

### External Access via Ingress

```yaml
ingress:
  enabled: true
  hosts:
    - grafana.yourdomain.com
  tls:
    - secretName: grafana-tls
      hosts:
        - grafana.yourdomain.com
```

---

## Useful Queries for Dashboards

### Application Performance

```promql
# Request rate by endpoint
sum(rate(http_requests_total[5m])) by (endpoint)

# Latency percentiles
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))  # p50
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))  # p95
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))  # p99

# Error rate percentage
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# Success rate
sum(rate(http_requests_total{status=~"2.."}[5m])) / sum(rate(http_requests_total[5m])) * 100
```

### Resource Usage

```promql
# CPU usage (requires kube-state-metrics)
rate(container_cpu_usage_seconds_total{pod=~"go-ping.*"}[5m])

# Memory usage
container_memory_usage_bytes{pod=~"go-ping.*"}

# Network I/O
rate(container_network_receive_bytes_total{pod=~"go-ping.*"}[5m])
rate(container_network_transmit_bytes_total{pod=~"go-ping.*"}[5m])
```

---

## Troubleshooting

### Data Source Not Working

```bash
# Check Prometheus is accessible from Grafana pod
kubectl exec -it <grafana-pod> -- wget -O- http://prometheus-server:80/-/healthy

# Check Grafana logs
kubectl logs -l app.kubernetes.io/name=grafana

# Verify service exists
kubectl get svc prometheus-server
```

### Dashboards Not Loading

```bash
# Check ConfigMap
kubectl get configmap grafana-dashboards -o yaml

# Check volume mounts
kubectl describe pod -l app.kubernetes.io/name=grafana
```

### Slow Queries

- Reduce time range
- Increase scrape interval
- Use recording rules in Prometheus
- Optimize PromQL queries (avoid `rate` over long ranges)

---

## Best Practices

### Dashboard Design

1. **Keep it simple** - Don't overcrowd with too many panels
2. **Use colors wisely** - Red for errors, green for success
3. **Add descriptions** - Help others understand the metrics
4. **Use variables** - Make dashboards reusable
5. **Group related panels** - Use rows to organize

### Query Optimization

1. **Use recording rules** for complex calculations
2. **Limit time ranges** - Don't query years of data
3. **Use appropriate intervals** - Match to scrape interval
4. **Avoid high cardinality** - Don't use too many label combinations

### Alerting

1. **Alert on symptoms, not causes** - Focus on user impact
2. **Avoid alert fatigue** - Don't alert on everything
3. **Set appropriate thresholds** - Based on SLOs
4. **Test alerts** - Verify they trigger correctly
5. **Document runbooks** - Add links in alert descriptions

---

## Learning Resources

### Essential Reading
- [Grafana Getting Started](https://grafana.com/docs/grafana/latest/getting-started/)
- [Dashboard Best Practices](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/best-practices/)
- [PromQL in Grafana](https://grafana.com/docs/grafana/latest/datasources/prometheus/)
- [Variables and Templating](https://grafana.com/docs/grafana/latest/dashboards/variables/)

### Video Tutorials
- [Grafana Crash Course](https://www.youtube.com/watch?v=Nkn5B8UYJ6M)
- [Creating Dashboards](https://www.youtube.com/watch?v=videoid)

### Dashboard Examples
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- Browse community dashboards for inspiration

---

## Future: Enhanced Observability with OpenTelemetry

Grafana works seamlessly with **OpenTelemetry** for advanced observability:

### What You Gain

**Current setup**:
- ✅ Metrics from Prometheus
- ❌ No distributed tracing
- ❌ Separate logging solution

**With OpenTelemetry**:
- ✅ Metrics from Prometheus (same as now)
- ✅ Distributed traces from Tempo/Jaeger
- ✅ Logs from Loki
- ✅ **Correlation** between all three!

### Grafana as Unified Observability Platform

```
Grafana Dashboard
├─ Metrics panel (Prometheus)
├─ Traces panel (Tempo)
└─ Logs panel (Loki)

Click a spike in metrics → See related traces → View associated logs
```

### Example: Debugging a Slow Request

1. **Metrics**: See p99 latency spike at 10:30 AM
2. **Click spike** → Jump to traces for that time
3. **Traces**: See request took 2.5s in database query
4. **Click trace ID** → See related logs
5. **Logs**: Find slow query warning

All in one Grafana dashboard!

### Additional Data Sources for Grafana

Beyond Prometheus, Grafana supports:

- **Tempo**: Distributed tracing (OpenTelemetry compatible)
- **Loki**: Log aggregation
- **Jaeger**: Alternative tracing backend
- **InfluxDB**: Time-series data
- **Elasticsearch**: Log search and analytics

### Migration Path

Your existing Grafana dashboards will continue working when adding OpenTelemetry:

1. Keep Prometheus data source
2. Add Tempo data source for traces
3. Add Loki data source for logs
4. Create unified dashboards with all three

No disruption to current metrics!

### Resources

- [Grafana Tempo](https://grafana.com/oss/tempo/) - Distributed tracing
- [Grafana Loki](https://grafana.com/oss/loki/) - Log aggregation
- [Correlations in Grafana](https://grafana.com/docs/grafana/latest/administration/correlations/)

## Next Steps

1. ✅ Complete Grafana installation
2. ✅ Create custom dashboards
3. ⏭️ Configure alerts
4. ⏭️ Add persistence for production
5. ⏭️ Set up user authentication
6. ⏭️ Export and version control dashboards
7. ⏭️ (Advanced) Add Tempo for distributed tracing
8. ⏭️ (Advanced) Add Loki for log aggregation

See [Monitoring README](../monitoring/README.md) for overall monitoring strategy.
