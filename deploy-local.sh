#!/bin/bash
# Deploy to local kind cluster

# Create kind cluster
kind create cluster --name task-management

# Load images (assuming built locally)
kind load docker-image your-registry/task-service:latest --name task-management
kind load docker-image your-registry/notification-service:latest --name task-management
kind load docker-image your-registry/gateway:latest --name task-management

# Install Helm charts
helm install task-service ./k8s/helm/task-service
helm install notification-service ./k8s/helm/notification-service
helm install gateway ./k8s/helm/gateway

# Port forward for testing
kubectl port-forward svc/gateway 8080:8080