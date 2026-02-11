# Kubernetes Tutorial Example

A minimal multi-service stack to demonstrate **ConfigMap**, **StatefulSet**, **multiple replicas**, **resource limits**, **HPA**, and **inter-service communication**.

## Architecture

| Service            | Language | Storage    | Role                                      |
|--------------------|----------|------------|-------------------------------------------|
| **go-service**       | Go       | PostgreSQL | Main API; stores items; calls other services |
| **python-redis**     | Python   | Redis      | Cache API (get/set); called by Go           |
| **java-mongo**       | Java (Spring Boot) | MongoDB | Document API (CRUD); called by Go   |

**Flow:** Client → Go service → (Postgres + Python-Redis + Java-Mongo). Go’s `/demo` endpoint calls both Python and Java services to show connectivity.

## Concepts Demonstrated

- **ConfigMap** – `app-config`: APP_NAME, PORT, DB names, cache TTL, etc.
- **Secret** – `app-secrets`: Postgres password (and other sensitive config).
- **StatefulSet** – Postgres, Redis, MongoDB: stable pod names (`postgres-0`, `redis-0`, `mongo-0`) and persistent volumes.
- **Deployment** – Go, Python Redis, and Java Mongo: multiple replicas, rolling updates.
- **Resource limits** – `requests`/`limits` on CPU and memory for all workloads.
- **Scaling** – Deployments run with `replicas: 2`; **HPA** scales Go and Python-Redis by CPU.

## Prerequisites

- Docker (to build images)
- Kubernetes cluster (minikube, kind, or cloud)
- `kubectl` configured for the cluster

For **HPA** to work, the cluster must have [metrics-server](https://github.com/kubernetes-sigs/metrics-server) installed (e.g. `minikube addons enable metrics-server`).

## Quick Start

### 1. Build images

From the repo root (e.g. `k8s-tutorial/`):

```bash
docker build -t go-service:latest ./go-service
docker build -t python-redis-service:latest ./python-redis
docker build -t python-mongo-service:latest ./python-mongo
```

**Minikube:** use the Docker daemon inside minikube so images are available in the cluster:

```bash
eval $(minikube docker-env)
docker build -t go-service:latest ./go-service
docker build -t python-redis-service:latest ./python-redis
docker build -t java-mongo-service:latest ./java-mongo
```

**Kind:** load images after building:

```bash
kind load docker-image go-service:latest python-redis-service:latest java-mongo-service:latest
```

### 2. Create namespace and base config

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
```

### 3. Deploy databases (StatefulSets)

Start Postgres, Redis, and MongoDB; wait until their pods are ready:

```bash
kubectl apply -f k8s/postgres-statefulset.yaml
kubectl apply -f k8s/redis-statefulset.yaml
kubectl apply -f k8s/mongo-statefulset.yaml
kubectl get pods -n k8s-tutorial -w
```

### 4. Deploy applications

```bash
kubectl apply -f k8s/go-deployment.yaml
kubectl apply -f k8s/python-redis-deployment.yaml
kubectl apply -f k8s/java-mongo-deployment.yaml
```

### 5. (Optional) NodePort and HPA

```bash
kubectl apply -f k8s/go-service-nodeport.yaml
kubectl apply -f k8s/hpa.yaml
```

### 6. Access the API

**Port-forward (works on any cluster):**

```bash
kubectl port-forward -n k8s-tutorial svc/go-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/demo
curl http://localhost:8080/items
curl -X POST http://localhost:8080/items -H "Content-Type: application/json" -d '{"name":"first item"}'
```

**NodePort (e.g. minikube):**

```bash
minikube service go-service-nodeport -n k8s-tutorial
# or: curl $(minikube ip):30080/health
```

## K8s Manifests Overview

| File                         | Purpose |
|-----------------------------|---------|
| `namespace.yaml`            | Namespace `k8s-tutorial` |
| `configmap.yaml`            | Non-sensitive app config |
| `secret.yaml`               | Passwords and secrets |
| `postgres-statefulset.yaml` | Postgres StatefulSet + headless Service + PVC |
| `redis-statefulset.yaml`    | Redis StatefulSet + headless Service + PVC |
| `mongo-statefulset.yaml`    | MongoDB StatefulSet + headless Service + PVC |
| `go-deployment.yaml`        | Go app Deployment (2 replicas) + Service |
| `python-redis-deployment.yaml` | Python Redis Deployment (2 replicas) + Service |
| `java-mongo-deployment.yaml`   | Java Mongo Deployment (2 replicas) + Service    |
| `hpa.yaml`                  | HPA for go-service and python-redis-service |
| `go-service-nodeport.yaml`  | NodePort for external access |

## Scaling

- **Manual:**  
  `kubectl scale deployment go-service -n k8s-tutorial --replicas=5`
- **HPA:**  
  After applying `hpa.yaml`, watch:  
  `kubectl get hpa -n k8s-tutorial`  
  Load the Go service (e.g. loop on `/health` or `/items`) to see replicas increase when CPU target is exceeded.

## Cleanup

```bash
kubectl delete namespace k8s-tutorial
```

This removes all resources in the namespace, including PVCs if the storage class allows.
