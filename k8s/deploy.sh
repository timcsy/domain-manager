#!/bin/bash

# Deploy script for Domain Manager on Kubernetes
set -e

# Configuration
# Use existing KUBECONFIG or default to ~/.kube/config
: ${KUBECONFIG:=~/.kube/config}

echo "🚀 Starting deployment of Domain Manager to Kubernetes..."
echo "📋 Using KUBECONFIG: ${KUBECONFIG}"
echo ""

# Check if we need to install NGINX Ingress Controller
echo "📦 Checking for Ingress Controller..."
if ! kubectl get ingressclass nginx &> /dev/null; then
    echo "⚠️  NGINX Ingress Controller not found"
    echo "Installing NGINX Ingress Controller..."
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

    echo "⏳ Waiting for Ingress Controller to be ready..."
    kubectl wait --namespace ingress-nginx \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=300s

    echo "✅ NGINX Ingress Controller installed successfully"
else
    echo "✅ NGINX Ingress Controller already installed"
fi

# Create namespace
echo "📁 Creating namespace..."
kubectl apply -f base/namespace.yaml

# Deploy RBAC
echo "🔐 Deploying RBAC..."
kubectl apply -f base/rbac.yaml

# Deploy PVC
echo "💾 Creating PersistentVolumeClaim..."
kubectl apply -f base/pvc.yaml

# Deploy Service
echo "🌐 Deploying Service..."
kubectl apply -f base/service.yaml

# Deploy Deployment
echo "🚢 Deploying Domain Manager..."
kubectl apply -f base/deployment.yaml

# Wait for deployment to be ready
echo "⏳ Waiting for Domain Manager to be ready..."
kubectl wait --namespace domain-manager \
  --for=condition=available deployment/domain-manager \
  --timeout=300s

# Deploy example service
echo "🎯 Deploying example service..."
kubectl apply -f base/example-service.yaml

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Current status:"
kubectl get pods -n domain-manager
kubectl get svc -n domain-manager
echo ""
echo "🔍 To check Domain Manager logs:"
echo "   kubectl logs -n domain-manager -l app=domain-manager -f"
echo ""
echo "🌐 To access Domain Manager, you can use port-forward:"
echo "   kubectl port-forward -n domain-manager svc/domain-manager 8080:80"
echo "   Then visit: http://localhost:8080"
echo ""
echo "📝 To deploy with Ingress (requires domain configuration):"
echo "   1. Update k8s/base/ingress.yaml with your domain"
echo "   2. kubectl apply -f k8s/base/ingress.yaml"
