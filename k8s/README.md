# Kubernetes Deployment Guide for Domain Manager

This directory contains Kubernetes manifests for deploying the Domain Manager application.

## Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured to access your cluster
- NGINX Ingress Controller (will be installed automatically if not present)
- (Optional) cert-manager for automatic TLS certificate management

## Quick Start

### 1. Set up KUBECONFIG

```bash
# Set your kubeconfig path
export KUBECONFIG=/path/to/your/kubeconfig.yaml
```

### 2. Deploy Domain Manager

```bash
cd k8s
chmod +x deploy.sh
./deploy.sh
```

This script will:
- Install NGINX Ingress Controller if needed
- Create the `domain-manager` namespace
- Deploy all necessary resources (RBAC, PVC, Service, Deployment)
- Deploy an example "hello-world" service for testing

### 2. Access Domain Manager

#### Option A: Port Forward (Recommended for testing)

```bash
kubectl port-forward -n domain-manager svc/domain-manager 8080:80
```

Then visit: http://localhost:8080

Default credentials:
- Username: `admin`
- Password: `admin`

#### Option B: Using Ingress (Production)

1. Update `k8s/base/ingress.yaml` with your domain name
2. Apply the Ingress:
   ```bash
   kubectl apply -f k8s/base/ingress.yaml
   ```
3. Configure DNS to point to your Ingress Controller's external IP

## File Structure

```
k8s/
├── base/
│   ├── namespace.yaml       # Namespace definition
│   ├── rbac.yaml            # ServiceAccount, Role, RoleBinding
│   ├── pvc.yaml             # PersistentVolumeClaim for database
│   ├── deployment.yaml      # Domain Manager deployment
│   ├── service.yaml         # Domain Manager service
│   ├── ingress.yaml         # Ingress for external access
│   └── example-service.yaml # Example service for testing
├── deploy.sh                # Deployment script
└── README.md                # This file
```

## Configuration

### Environment Variables

The Domain Manager deployment uses these environment variables:

- `PORT`: HTTP server port (default: 8080)
- `DB_PATH`: SQLite database file path (default: /data/domain-manager.db)
- `K8S_IN_CLUSTER`: Enable in-cluster Kubernetes API access (default: true)

### Storage

The deployment uses a PersistentVolumeClaim with:
- Size: 1Gi
- Access Mode: ReadWriteOnce
- StorageClass: `vultr-block-storage` (modify in `pvc.yaml` if needed)

### Resource Limits

Default resource configuration:
- Requests: 100m CPU, 128Mi memory
- Limits: 500m CPU, 512Mi memory

## Testing Domain Manager

### 1. Check Deployment Status

```bash
# Check pods
kubectl get pods -n domain-manager

# Check services
kubectl get svc -n domain-manager

# View logs
kubectl logs -n domain-manager -l app=domain-manager -f
```

### 2. Test with Example Service

The deployment includes a "hello-world" service in the default namespace:

```bash
# Check example service
kubectl get svc -n default hello-world
```

Use Domain Manager UI to:
1. Log in with admin credentials
2. Navigate to "Domains" section
3. Create a new domain mapping:
   - Domain Name: `hello.example.com`
   - Target Service: `hello-world`
   - Target Namespace: `default`
   - Target Port: `80`
4. Domain Manager will automatically create a Kubernetes Ingress

### 3. Verify Ingress Creation

```bash
# List ingresses created by Domain Manager
kubectl get ingress -A

# Check specific ingress details
kubectl describe ingress -n default domain-1
```

## Troubleshooting

### Pod not starting

```bash
# Check pod events
kubectl describe pod -n domain-manager -l app=domain-manager

# Check logs
kubectl logs -n domain-manager -l app=domain-manager
```

### RBAC permissions issues

```bash
# Verify ServiceAccount exists
kubectl get serviceaccount -n domain-manager

# Check ClusterRole and ClusterRoleBinding
kubectl get clusterrole domain-manager
kubectl get clusterrolebinding domain-manager
```

### Storage issues

```bash
# Check PVC status
kubectl get pvc -n domain-manager

# Check available storage classes
kubectl get storageclass
```

## Uninstall

To remove Domain Manager:

```bash
# Delete all Domain Manager resources
kubectl delete namespace domain-manager

# Delete ClusterRole and ClusterRoleBinding
kubectl delete clusterrole domain-manager
kubectl delete clusterrolebinding domain-manager

# (Optional) Delete example service
kubectl delete -f k8s/base/example-service.yaml
```

## Production Considerations

### 1. Use Image from GitHub Container Registry

The deployment uses: `ghcr.io/timcsy/domain-manager:latest`

For production, use specific version tags:
```yaml
image: ghcr.io/timcsy/domain-manager:v1.0.0
```

### 2. Configure TLS with cert-manager

Install cert-manager first:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

Create ClusterIssuer for Let's Encrypt:
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

### 3. Set up Monitoring

Add monitoring labels and Prometheus annotations:
```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

### 4. Database Backup

Consider setting up regular backups of the SQLite database:
```bash
# Create a backup cronjob
kubectl create cronjob domain-manager-backup \
  --image=busybox \
  --schedule="0 2 * * *" \
  -- /bin/sh -c "cp /data/domain-manager.db /backup/domain-manager-$(date +%Y%m%d).db"
```

## Support

For issues and questions:
- GitHub Issues: https://github.com/timcsy/domain-manager/issues
- Documentation: https://github.com/timcsy/domain-manager
