# Helm Chart - Go Ping Application

This Helm chart deploys a simple Go-based ping service that responds with "pong" to HTTP requests. It is designed for educational purposes to demonstrate cloud-native application deployment using Kubernetes and Helm.

The go-ping application is developed locally and containerized using Docker and is located in the `/api` directory.

## Features

- ✅ Lightweight Go HTTP service with Chi router
- ✅ Health and readiness probes
- ✅ Structured logging (ECS-compliant)
- ✅ Graceful shutdown handling
- ✅ Production-ready security contexts
- ✅ Configurable via environment variables
- ✅ Optional autoscaling (HPA)
- ✅ Optional Ingress support

## Prerequisites

- Kubernetes cluster (local or cloud-based)
- Helm 3.x installed
- kubectl configured to interact with your Kubernetes cluster
- Docker installed for building container images
- (Optional) KinD for local Kubernetes cluster setup

## Quick Start

### 1. Build and Push Docker Image (Optional)

If using a custom image:

```bash
# Navigate to the API directory
cd api

# Build the Docker image
docker build -t ghcr.io/arifjehoh/orchestrated-ping:latest .

# Push to registry (requires authentication)
docker push ghcr.io/arifjehoh/orchestrated-ping:latest
```

### 2. Install the Chart

**Basic installation:**

```bash
# From the project root
helm install go-ping ./charts/go-ping
```

**With custom values:**

```bash
helm install go-ping ./charts/go-ping \
  --set image.tag=v1.0.0 \
  --set env.environment=production \
  --set replicaCount=3
```

**Using a values file:**

```bash
helm install go-ping ./charts/go-ping -f my-values.yaml
```

### 3. Verify Installation

```bash
# Check deployment status
kubectl get pods -l app.kubernetes.io/name=go-ping

# Check service
kubectl get svc go-ping

# View logs
kubectl logs -l app.kubernetes.io/name=go-ping --tail=50
```

## Configuration

### Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Container image repository | `ghcr.io/arifjehoh/orchestrated-ping` |
| `image.tag` | Container image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `replicaCount` | Number of replicas | `1` |
| `env.environment` | Application environment | `production` |
| `env.port` | Server port | `8080` |
| `env.readTimeout` | HTTP read timeout | `15s` |
| `env.writeTimeout` | HTTP write timeout | `15s` |
| `env.shutdownTimeout` | Graceful shutdown timeout | `30s` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `ingress.enabled` | Enable ingress | `false` |
| `autoscaling.enabled` | Enable HPA | `false` |
| `resources.requests.cpu` | CPU request | `20m` |
| `resources.requests.memory` | Memory request | `20Mi` |
| `resources.limits.cpu` | CPU limit | `20m` |
| `resources.limits.memory` | Memory limit | `20Mi` |

### Full Configuration

See [values.yaml](values.yaml) for all available configuration options.

## Usage

### Accessing the Service

**From within the cluster:**

```bash
# Port-forward to access locally
kubectl port-forward svc/go-ping 8080:8080

# Test the endpoints
curl http://localhost:8080/ping
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

**With LoadBalancer service:**

```bash
# Change service type
helm upgrade go-ping ./charts/go-ping --set service.type=LoadBalancer

# Get external IP
kubectl get svc go-ping

# Access the service
curl http://<EXTERNAL-IP>:8080/ping
```

**With Ingress:**

```bash
# Update values.yaml or use --set
helm upgrade go-ping ./charts/go-ping --set ingress.enabled=true \
  --set ingress.hosts[0].host=ping.example.com \
  --set ingress.hosts[0].paths[0].path=/ \
  --set ingress.hosts[0].paths[0].pathType=Prefix

# Access via hostname
curl http://ping.example.com/ping
```

### API Endpoints

| Endpoint | Method | Description | Response |
|----------|--------|-------------|----------|
| `/ping` | GET | Ping endpoint | `{"status":"success","message":"pong","time":"..."}` |
| `/health` | GET | Health check | `{"status":"healthy","uptime":"..."}` |
| `/ready` | GET | Readiness check | `{"status":"ready","message":"...","time":"..."}` |

### Example Responses

**Ping:**
```json
{
  "status": "success",
  "message": "pong",
  "time": "2025-12-24T10:30:00Z"
}
```

**Health:**
```json
{
  "status": "healthy",
  "uptime": "2h15m30s"
}
```

**Ready:**
```json
{
  "status": "ready",
  "message": "application is ready to serve traffic",
  "time": "2025-12-24T10:30:00Z"
}
```

## Scaling

### Manual Scaling

```bash
# Scale to 5 replicas
kubectl scale deployment go-ping --replicas=5

# Or via Helm
helm upgrade go-ping ./charts/go-ping --set replicaCount=5
```

### Autoscaling (HPA)

```bash
# Enable autoscaling
helm upgrade go-ping ./charts/go-ping \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=2 \
  --set autoscaling.maxReplicas=10 \
  --set autoscaling.targetCPUUtilizationPercentage=80
```

## Production Configuration

Example production values:

```yaml
# production-values.yaml
replicaCount: 3

env:
  environment: production
  readTimeout: 30s
  writeTimeout: 30s

resources:
  requests:
    cpu: 100m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 128Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: ping.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: go-ping-tls
      hosts:
        - ping.yourdomain.com

podDisruptionBudget:
  enabled: true
  minAvailable: 2
```

Deploy:
```bash
helm install go-ping ./charts/go-ping -f production-values.yaml
```

## Upgrading

```bash
# Upgrade with new image version
helm upgrade go-ping ./charts/go-ping --set image.tag=v1.1.0

# Upgrade with new values file
helm upgrade go-ping ./charts/go-ping -f new-values.yaml

# View upgrade history
helm history go-ping

# Rollback to previous version
helm rollback go-ping
```

## Uninstalling

```bash
# Uninstall the release
helm uninstall go-ping

# Verify removal
kubectl get all -l app.kubernetes.io/name=go-ping
```

## Troubleshooting

### Pod not starting

```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/name=go-ping

# Describe pod for events
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>
```

### Service not accessible

```bash
# Verify service endpoints
kubectl get endpoints go-ping

# Test from within cluster
kubectl run curl --image=curlimages/curl -it --rm -- curl http://go-ping:8080/ping
```

### Health probe failures

```bash
# Check probe configuration
kubectl get pod <pod-name> -o yaml | grep -A 10 livenessProbe

# Test health endpoint
kubectl port-forward <pod-name> 8080:8080
curl http://localhost:8080/health
```

## Integration with NGINX Proxy

This chart can optionally work with the NGINX proxy chart for advanced routing. See [nginx-proxy README](../nginx-proxy/README.md) for details.

```bash
# Deploy go-ping first
helm install go-ping ./charts/go-ping

# Then deploy NGINX proxy
helm install nginx-proxy ./charts/nginx-proxy
```

## Development

### Local Testing with KinD

```bash
# Create local cluster
kind create cluster --name go-ping-test

# Load image into KinD
kind load docker-image ghcr.io/arifjehoh/orchestrated-ping:latest --name go-ping-test

# Install chart
helm install go-ping ./charts/go-ping --set image.pullPolicy=Never

# Test
kubectl port-forward svc/go-ping 8080:8080
```

### Linting and Validation

```bash
# Lint the chart
helm lint ./charts/go-ping

# Dry-run installation
helm install go-ping ./charts/go-ping --dry-run --debug

# Template rendering
helm template go-ping ./charts/go-ping > rendered.yaml
```

## License

See [LICENSE](../../LICENSE) file in the project root.

## Contributing

This is an educational project demonstrating cloud-native deployment patterns.