# BytePort Consul Service Mesh Configuration
# Deploy via: consul config write contrib/consul/service-defaults.hcl

kind      = "service-defaults"
name      = "byteport"
protocol  = "http"
mesh      = true

# Health check configuration
checks = [
  {
    name     = "BytePort API Health"
    http     = "http://127.0.0.1:8080/healthz"
    interval = "10s"
    timeout  = "2s"
  },
  {
    name     = "BytePort MCP Health"
    http     = "http://127.0.0.1:8081/healthz"
    interval = "15s"
    timeout  = "2s"
  }
]

# Circuit breaker configuration (compatible with W87 Go implementation)
expose = {}

# Rate limiting defaults
limits = {
  max_connections        = 100
  max_pending_requests   = 1000
  max_requests_per_second = 500
}

# Upstream timeouts
connect {
  sidecar_service {
    proxy {
      upstreams = [
        {
          destination_name = "postgres"
          local_bind_port  = 5432
        },
        {
          destination_name = "prometheus"
          local_bind_port  = 9090
        }
      ]
    }
  }
}
