# Local Docker Registry Setup for KinD

This guide explains how to set up a local Docker registry for use with KinD (Kubernetes in Docker). This is an **optional** configuration that was skipped in favor of using GitHub Container Registry (ghcr.io).

## Why Use a Local Registry?

### Benefits
- ✅ **Offline development** - No internet required for image pulls
- ✅ **Faster iteration** - No push/pull from remote registry
- ✅ **Air-gapped environments** - Works behind corporate firewalls
- ✅ **Testing CI/CD** - Simulate registry workflows locally
- ✅ **Cost optimization** - Avoid bandwidth charges during development
- ✅ **Privacy** - Images stay on your machine

### When NOT Needed
- ❌ Already using a cloud registry (ghcr.io, Docker Hub, ECR, etc.)
- ❌ Learning production-like workflows (cloud registries are more realistic)
- ❌ Multi-architecture builds (easier with GitHub Actions)
- ❌ Team collaboration (shared registry is better)

## Prerequisites

- Docker installed and running
- KinD CLI installed
- kubectl installed

## Setup Instructions

### Step 1: Create Local Registry Container

Create a Docker registry running on your local machine:

```bash
# Start local Docker registry on port 5001
docker run -d --restart=always \
  -p 5001:5000 \
  --name kind-registry \
  registry:2

# Verify registry is running
docker ps | grep kind-registry
curl http://localhost:5001/v2/_catalog
```

**Why port 5001?** Port 5000 is often used by macOS AirPlay, so we map to 5001 externally.

### Step 2: Create KinD Cluster with Registry Configuration

Create a KinD cluster configured to use the local registry:

```bash
# Create cluster configuration file
cat <<EOF > kind-with-registry.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5001"]
    endpoint = ["http://kind-registry:5000"]
nodes:
- role: control-plane
- role: worker  # Optional: add worker nodes
EOF

# Create the cluster
kind create cluster --config=kind-with-registry.yaml --name dev

# Connect registry to kind network
docker network connect kind kind-registry

# Verify connection
docker exec kind-control-plane curl http://kind-registry:5000/v2/_catalog
```

### Step 3: Document the Registry for KinD

Create a ConfigMap so KinD knows about the local registry:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:5001"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF
```

### Step 4: Build and Push Images to Local Registry

```bash
# Build the image
cd api
docker build -t localhost:5001/orchestrated-ping:latest .

# Push to local registry
docker push localhost:5001/orchestrated-ping:latest

# Verify image is in registry
curl http://localhost:5001/v2/_catalog
curl http://localhost:5001/v2/orchestrated-ping/tags/list
```

### Step 5: Update Helm Values for Local Registry

Update your Helm chart values to use the local registry:

```yaml
# values-local.yaml
image:
  repository: localhost:5001/orchestrated-ping
  tag: latest
  pullPolicy: IfNotPresent
```

Deploy using local values:

```bash
helm install go-ping ./charts/go-ping -f values-local.yaml
```

## Complete Example Workflow

### Initial Setup (One-time)

```bash
# 1. Start local registry
docker run -d --restart=always -p 5001:5000 --name kind-registry registry:2

# 2. Create KinD cluster with registry config
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5001"]
    endpoint = ["http://kind-registry:5000"]
nodes:
- role: control-plane
EOF

# 3. Connect registry to kind network
docker network connect kind kind-registry

# 4. Add registry ConfigMap
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:5001"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF
```

### Daily Development Workflow

```bash
# 1. Make code changes
cd api
# ... edit files ...

# 2. Build and push to local registry
docker build -t localhost:5001/orchestrated-ping:latest .
docker push localhost:5001/orchestrated-ping:latest

# 3. Update deployment (force pull new image)
kubectl rollout restart deployment go-ping
# OR reinstall with Helm
helm upgrade go-ping ./charts/go-ping -f values-local.yaml

# 4. Verify deployment
kubectl get pods
kubectl logs -l app.kubernetes.io/name=go-ping
```

## Alternative: Load Images Directly into KinD

If you don't want a registry, you can load images directly:

```bash
# Build image
docker build -t orchestrated-ping:latest ./api

# Load directly into KinD
kind load docker-image orchestrated-ping:latest

# Use in Helm with Never pull policy
helm install go-ping ./charts/go-ping \
  --set image.repository=orchestrated-ping \
  --set image.tag=latest \
  --set image.pullPolicy=Never
```

**Pros**: No registry needed, simpler setup  
**Cons**: Must reload image for every change, doesn't work with multiple clusters

## Troubleshooting

### Registry Not Accessible from KinD

```bash
# Check if registry container is running
docker ps | grep kind-registry

# Check if registry is on kind network
docker network inspect kind | grep kind-registry

# Test from KinD node
docker exec kind-control-plane curl http://kind-registry:5000/v2/_catalog
```

### Image Pull Errors

```bash
# Verify image exists in registry
curl http://localhost:5001/v2/orchestrated-ping/tags/list

# Check pod events
kubectl describe pod <pod-name>

# Verify image name matches exactly
kubectl get pod <pod-name> -o jsonpath='{.spec.containers[0].image}'
```

### Registry Data Persistence

The registry stores data in a Docker volume. To persist across restarts:

```bash
# Create named volume
docker volume create kind-registry-data

# Start registry with volume
docker run -d --restart=always \
  -p 5001:5000 \
  --name kind-registry \
  -v kind-registry-data:/var/lib/registry \
  registry:2
```

## Cleanup

```bash
# Stop and remove registry
docker stop kind-registry
docker rm kind-registry

# Remove registry data volume (optional)
docker volume rm kind-registry-data

# Delete KinD cluster
kind delete cluster --name dev
```

## Production Considerations

**Do NOT use a local registry for production!**

Use managed registries instead:
- **GitHub Container Registry** (ghcr.io) - Free for public repos
- **Docker Hub** - Free tier available
- **AWS ECR** - Integrated with EKS
- **Google Container Registry** (gcr.io) - Integrated with GKE
- **Azure Container Registry** - Integrated with AKS
- **Harbor** - Self-hosted enterprise registry

## References

- [KinD Local Registry Documentation](https://kind.sigs.k8s.io/docs/user/local-registry/)
- [Docker Registry Documentation](https://docs.docker.com/registry/)
- [containerd Registry Configuration](https://github.com/containerd/containerd/blob/main/docs/cri/registry.md)

## Why This Was Skipped

For this project, we use **GitHub Container Registry (ghcr.io)** because:

1. **Production-like workflow** - Mirrors real-world deployments
2. **GitHub Actions integration** - Automatic builds on push
3. **Multi-architecture support** - ARM64 + AMD64 in one manifest
4. **No local setup required** - Works immediately
5. **Team collaboration** - Sharable across developers
6. **Learning value** - Better represents cloud-native practices

The local registry is excellent for specific use cases, but ghcr.io provides a better learning experience for cloud-native development patterns.
