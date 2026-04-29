#!/usr/bin/env bash
set -euo pipefail

# Installs ingress-nginx into k3s.

kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.1/deploy/static/provider/cloud/deploy.yaml

echo "Waiting for ingress-nginx controller to be ready..."
kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller --timeout=180s

kubectl -n ingress-nginx patch svc ingress-nginx-controller --type merge -p '{
  "spec": {
    "type": "NodePort",
    "externalTrafficPolicy": "Local",
    "ports": [
      {"name":"http","port":80,"protocol":"TCP","targetPort":"http","nodePort":80},
      {"name":"https","port":443,"protocol":"TCP","targetPort":"https","nodePort":443}
    ]
  }
}'
