# NGINX Proxy Helm Chart

NGINX reverse proxy for the orchestrated-ping service. This chart is **optional** - the go-ping application works standalone without it.

## Purpose

Provides a dedicated NGINX reverse proxy that sits in front of the go-ping service for:
- Load balancing
- SSL/TLS termination (when configured)
- Request routing and rewriting
- Advanced caching
- Rate limiting
- Custom NGINX configurations

## Usage

### Standalone go-ping (No NGINX)

Deploy just the go-ping application:

```bash
helm install go-ping ./charts/go-ping
```

Access it directly via the go-ping service.

### With NGINX Proxy

1. First, deploy go-ping:
```bash
helm install go-ping ./charts/go-ping
```

2. Then deploy NGINX proxy:
```bash
helm install nginx-proxy ./charts/nginx-proxy
```

NGINX will automatically proxy to the `go-ping` service.

### Configuration

Key values in `values.yaml`:

```yaml
backend:
  serviceName: go-ping        # Name of the backend service
  servicePort: 8080           # Port of the backend service
  namespace: ""               # Leave empty for same namespace

service:
  type: LoadBalancer          # Expose NGINX externally
  port: 80
```

### Cross-Namespace Deployment

If go-ping is in a different namespace:

```yaml
backend:
  serviceName: go-ping
  servicePort: 8080
  namespace: production       # Namespace where go-ping is deployed
```

## Architecture

```
Internet → LoadBalancer → NGINX Pods → go-ping Service → go-ping Pods
```

## Endpoints

All endpoints are proxied from go-ping:
- `GET /ping` - Ping endpoint
- `GET /health` - Health check
- `GET /ready` - Readiness check

## Installation Examples

### Basic installation:
```bash
helm install nginx-proxy ./charts/nginx-proxy
kubectl port-forward svc/nginx-proxy 8081:80 &
```

### With custom backend:
```bash
helm install nginx-proxy ./charts/nginx-proxy \
  --set backend.serviceName=my-service \
  --set backend.servicePort=3000
```

### With NodePort instead of LoadBalancer:
```bash
helm install nginx-proxy ./charts/nginx-proxy \
  --set service.type=NodePort
```

## Uninstalling

```bash
helm uninstall nginx-proxy
```

The go-ping service will continue to work independently.
