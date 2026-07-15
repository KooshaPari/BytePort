# Grafana Dashboards — BytePort

Three production-ready dashboards provisioned via the Grafana sidecar provider.

## Dashboards

| File | UID | Purpose |
|------|-----|---------|
| `dashboards/byteport-overview.json` | `byteport-ops` | P95 latency, success rate, active deployments, $/hr spend, error table |
| `dashboards/byteport-procurement.json` | `byteport-sla` | Procurement queue depth, network egress |
| `dashboards/byteport-errors.json` | `byteport-errors` | 5xx rate, rate-limit hits |

## Provisioning

`provisioning/dashboards/byteport.yaml` and `provisioning/datasources/prometheus.yaml` wire dashboards and the Prometheus datasource automatically. Mount `/etc/grafana/dashboards` to `/etc/grafana/provisioning/dashboards` (see `docker-compose.yml` profile `observability`).

## Metrics Source

All metrics emitted by `backend/internal/infrastructure/observability/metrics.go` and the `/metrics` endpoint documented in the backend observability README.
