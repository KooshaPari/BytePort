# BytePort Consul Connect ingress gateway for external traffic.
# Deploy via: consul config write contrib/consul/ingress-gateway.hcl

kind = "ingress-gateway"
name = "byteport-ingress"

listeners = [
  {
    port     = 8080
    protocol = "http"

    services = [
      {
        name  = "byteport"
        hosts = ["api.byteport.dev"]
      }
    ]
  },
  {
    port     = 8081
    protocol = "http"

    services = [
      {
        name  = "byteport"
        hosts = ["mcp.byteport.dev"]
      }
    ]
  }
]
