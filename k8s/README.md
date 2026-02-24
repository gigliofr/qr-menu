# Kubernetes Deployment

## Apply

kubectl apply -k ./k8s

## Update image

- Build and load the image locally, or push to a registry and update deployment.yaml.

## Verify

kubectl get pods -n qr-menu
kubectl get svc -n qr-menu
