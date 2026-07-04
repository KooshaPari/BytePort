---
sidebar_position: 4
---

# Recipe — Monitoring hook → Grafana

## What you'll build

Push BytePort's per-deploy metrics (CPU, memory, request rate,
p99 latency, error rate) into a self-hosted Grafana.

**Time: ~12 minutes.**

## Prereqs

- A running Grafana 10+ with a Prometheus data source (any kind)
- Network egress from your Grafana instance to your BytePort region
- Owner access to your BytePort workspace

## Steps

```bash
# 1. Generate a metrics API key
byteport metrics keys add grafana \
  --workspace $WORKSPACE \
  --scopes read:metrics

# 2. Configure BytePort to expose Prometheus-format metrics
cat <<'TOML' > .byteport/metrics.toml
[metrics]
exposition = "prometheus"
push       = "https://grafana.example.com/api/v1/push"
push_format = "remote_write"
TOML

# 3. Wire the remote_write endpoint on Grafana
# In Grafana → Connections → Data sources → Add Prometheus:
#   URL:  https://byteport.example.com/api/v1/metrics
#   Auth: Bearer <metrics-API-key>
#   Scrape interval: 15s

# 4. Import the BytePort dashboard
#    Grafana → Dashboards → Import → paste dashboard ID 24000 (the
#    official "BytePort production" dashboard)
```

## Verify

In Grafana → Explore:
- `sum(byteport_deploys_total{workspace="..."})` should show a counter
- `byteport_request_duration_seconds_bucket{le="0.1"}` should show p99

## Cleanup

```bash
byteport metrics keys revoke grafana
```

## Related

- [Backup Postgres to S3](postgres-backup-to-s3.md) — same Vault-secret pattern for the metrics key
- Cookbook category: **Observability**
