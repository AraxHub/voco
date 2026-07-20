## Monitoring (лёгкий стек: Prometheus + Grafana)

Цель: базовые метрики (**ingress-nginx**, node-exporter, kube-state-metrics), без публикации UI в интернет. Доступ — `kubectl port-forward` и при необходимости SSH `-L`.

---

### Сначала: Grafana

Конфиг лежит в `deployment/monitoring/values-grafana.yaml`: там **provisioning источника Prometheus** (`uid: prometheus`) и три импортных дашборда в папке **Voco** (скачаются init-контейнером с grafana.com).

**На VPS после `git pull`:**

```bash
cd ~/voco
chmod +x deployment/monitoring/*.sh

helm repo add grafana https://grafana.github.io/helm-charts 2>/dev/null || true
helm repo update
helm upgrade --install grafana grafana/grafana \
  -n monitoring \
  -f deployment/monitoring/values-grafana.yaml

kubectl -n monitoring rollout status deploy/grafana --timeout=300s
kubectl -n monitoring get pods -l app.kubernetes.io/name=grafana -o wide
```

Пока под в статусе `Init`, идёт контейнер **`download-dashboards`** (качает JSON). Если `Running` уже долго, но без дашбордов — лог:

```bash
kubectl -n monitoring logs deploy/grafana -c download-dashboards
```

Убедиться, что в ConfigMap попал provisioning датасорса:

```bash
kubectl -n monitoring get cm grafana -o yaml | grep -A25 "datasources.yaml:"
```

**Пароль `admin`:**

```bash
kubectl -n monitoring get secret grafana -o jsonpath='{.data.admin-password}' | base64 -d; echo
```

**Доступ к UI:** на VPS в отдельном терминале (процесс не закрывать):

```bash
./deployment/monitoring/port-forward.sh
```

С ноутбука (пока живой `port-forward` на VPS):

```bash
ssh -N -L 3000:127.0.0.1:3000 root@<VPS_IP>
```

Браузер: `http://localhost:3000` → вход `admin` + пароль из Secret.

В интерфейсе Grafana 10+ источники: **Connections → Data sources** (или **Connections → Prometheus** сразу). Должен быть **Prometheus** с URL вида `http://prometheus-server`, **Save & test** — OK.

Дашборды: **Dashboards → Browse** → папка **Voco** (Node Exporter, NGINX Ingress, Kubernetes workloads).

**Если источника нет после helm upgrade:** чаще всего не переустановился релиз с `-f deployment/monitoring/values-grafana.yaml` или образ старый без пересборки конфигмапа — повторить `helm upgrade` и `kubectl rollout restart deploy/grafana -n monitoring`.

---

### Общий install (Prometheus + Grafana сразу)

```bash
cd ~/voco
chmod +x deployment/monitoring/*.sh
./deployment/monitoring/install.sh
```

На VPS для UI (Grafana `:3000`, Prometheus `:9090` локально):

```bash
./deployment/monitoring/port-forward.sh
```

С ноутбука:

```bash
ssh -N -L 3000:127.0.0.1:3000 -L 9090:127.0.0.1:9090 root@<VPS_IP>
```

Открыть `http://localhost:3000` и `http://localhost:9090`.

### Что ставим кратко

- **Prometheus** (`prometheus-community/prometheus`): alertmanager/pushgateway в наших values выключены.
- **Grafana** (`grafana/grafana`): ClusterIP, provisioning из `values-grafana.yaml`.

### Скрейп-таргеты

В `values-prometheus.yaml` включены extra scrape jobs:

- `ingress-nginx` — поды controller, `/metrics` на **:10254**, target-лейблы `namespace`/`pod` для дашбордов. Контроллеру нужен аргумент **`--enable-metrics=true`** (в официальном cloud-manifest v1.12.1 его нет → на :10254 только `go_*`; см. `deployment/k3s/ingress-install.sh`).
- `livekit` — под в `voco`, порт **6789** (`prometheus` в манифесте), не путать с signaling **7880**

Если какие-то job’ы не “зелёные”, смотри:

```bash
kubectl -n ingress-nginx get pods -l app.kubernetes.io/component=controller
kubectl -n voco get svc livekit-svc -o yaml
```

