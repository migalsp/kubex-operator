# Kubex Installation Guide

## Target Audience
This guide is intended for **DevOps Engineers, Platform Engineers, and Cluster Administrators** who want to deploy the Kubex Operator to manage and optimize their Kubernetes environments.

Kubex is designed to be as simple to install and maintain as possible, leveraging standard Helm and OCI artifacts.

---

## Prerequisites

Before installing Kubex, ensure you have:
1. **Kubernetes Cluster** (v1.22+ recommended).
2. **Helm** (v3.13+ installed locally).
3. **Metrics Server**: Ensure the Kubernetes Metrics Server is installed in your cluster (required for Namespace Optimization and Cluster Insights).
    ```bash
    kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
    ```

---

## Installation Strategy

Kubex is distributed as an OCI-compliant Helm chart hosted directly on GitHub Container Registry (GHCR).

### Step 1: Create Namespace
It is highly recommended to install Kubex in its own dedicated namespace.
```bash
kubectl create namespace kubex
```

### Step 2: Install via Helm
Install the operator directly from the OCI registry. This eliminates the need to manually add `.tgz` repositories to your local Helm cache.

```bash
helm upgrade --install kubex-operator oci://ghcr.io/migalsp/kubex-operator \
  --version v1.0.0 \
  --namespace kubex
```

*(Note: Replace `v1.0.0` with the latest release tag found on the GitHub Releases page).*

### Step 3: Verify Deployment
Ensure the operator pod is running and healthy:
```bash
kubectl get pods -n kubex
# NAME                              READY   STATUS    RESTARTS   AGE
# kubex-operator-5b8cb4b8b6-x4jz2   1/1     Running   0          45s
```

---

## Custom Configuration (values.yaml)

Kubex ships with sane defaults, requiring minimal resource requests (100m CPU / 128Mi RAM). However, you can deeply customize the deployment.

To view default values:
```bash
helm show values oci://ghcr.io/migalsp/kubex-operator --version v1.0.0
```

**Common Overrides (`my-values.yaml`):**
```yaml
# deploy/helm/kubex-operator/values.yaml

replicaCount: 1

image:
  repository: ghcr.io/migalsp/kubex-operator
  pullPolicy: IfNotPresent
  tag: "v1.0.0"

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi
```

Apply your overrides during installation:
```bash
helm upgrade --install kubex-operator oci://ghcr.io/migalsp/kubex-operator \
  --version v1.0.0 \
  --namespace kubex \
  -f my-values.yaml
```

---

## Exposing the UI Dashboard

The Kubex Operator provides a stunning real-time React dashboard. By default, it is exposed as a `ClusterIP` service to prevent unauthorized external access.

**Option A: Local Port-Forwarding (Recommended for quick view)**
```bash
kubectl port-forward svc/kubex-operator 8082:8082 -n kubex
```
*Open `http://localhost:8082` in your browser.*

**Option B: Ingress Configuration (For persistent team access)**
Create an Ingress resource to route traffic to the `kubex-operator` service on port `8082`. Ensure you secure this route with appropriate authentication (e.g., OAuth2 Proxy or an internal VPN).

---

## Upgrading

When a new version of Kubex is released, upgrading is seamless using Helm:
```bash
helm upgrade kubex-operator oci://ghcr.io/migalsp/kubex-operator \
  --version v1.1.0 \
  --namespace kubex
```

## Uninstalling

To completely remove the operator and all its components (Note: this **will not** delete namespaces, but it **will** delete the scaling and optimization CRDs applied to the cluster):

```bash
helm uninstall kubex-operator -n kubex
kubectl delete namespace kubex
```
