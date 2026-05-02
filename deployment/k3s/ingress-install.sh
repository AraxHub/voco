#!/usr/bin/env bash
set -euo pipefail

# Installs ingress-nginx into k3s.

kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.1/deploy/static/provider/cloud/deploy.yaml

echo "Waiting for ingress-nginx controller to be ready..."
kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller --timeout=180s

# В controller-v1.12.1 у бинари default для --enable-metrics = false (см. pkg/flags).
# Облачный deploy.yaml этого флага не задаёт → на :10254 только go_*; префикса
# nginx_ingress_controller_* нет → Grafana 9614 пустая. Включаем явно:
DEPLOY_ARGS="$(kubectl -n ingress-nginx get deploy ingress-nginx-controller -o jsonpath='{.spec.template.spec.containers[0].args}' 2>/dev/null || echo '')"
if printf '%s' "$DEPLOY_ARGS" | grep -q -- '--enable-metrics=false'; then
  echo "WARNING: ingress-nginx-controller уже с --enable-metrics=false; убери флаг через kubectl edit, иначе метрик ingress не будет." >&2
elif ! printf '%s' "$DEPLOY_ARGS" | grep -q -- '--enable-metrics=true'; then
  echo "Patching ingress-nginx-controller: add --enable-metrics=true"
  kubectl -n ingress-nginx patch deployment ingress-nginx-controller --type='json' -p='[
    {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--enable-metrics=true"}
  ]'
  kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller --timeout=180s
fi

# In k3s we disabled servicelb, so the default ingress-nginx Service of type LoadBalancer
# won't expose ports 80/443 on the node. Switch it to NodePort and pin to 80/443.
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
