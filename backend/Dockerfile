# =============================================================================
# BytePort backend — multi-stage Dockerfile
#
# Build targets (use --target):
#   backend    — Go API server (default)
#   mcp-server — MCP stdio server
#   dev        — development image with hot-reload tools
# =============================================================================
# Stage 1: Go build environment
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Cache module downloads
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source
COPY backend/ .

# Build the API server binary
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/byteport ./main.go

# Build the MCP server binary (optional, separate target)
FROM builder AS mcp-builder

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mcp-server ./cmd/mcp-server/

# =============================================================================
# Stage 2: API server runtime
FROM gcr.io/distroless/static-debian12:nonroot AS backend

LABEL org.opencontainers.image.title="BytePort API"
LABEL org.opencontainers.image.description="BytePort backend API server"
LABEL org.opencontainers.image.source="https://github.com/kooshapari/byteport"
LABEL org.opencontainers.image.licenses="MIT"

COPY --from=builder /out/byteport /byteport

EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/byteport"]

# =============================================================================
# Stage 3: MCP server runtime
FROM gcr.io/distroless/static-debian12:nonroot AS mcp-server

LABEL org.opencontainers.image.title="BytePort MCP Server"
LABEL org.opencontainers.image.description="BytePort MCP stdio server for agent-readiness"
LABEL org.opencontainers.image.source="https://github.com/kooshapari/byteport"
LABEL org.opencontainers.image.licenses="MIT"

COPY --from=mcp-builder /out/mcp-server /mcp-server

USER 65532:65532

ENTRYPOINT ["/mcp-server"]

# =============================================================================
# Stage 4: Development image
FROM golang:1.25-alpine AS dev

RUN apk add --no-cache git curl make

WORKDIR /app

# Install hot-reload tool
RUN go install github.com/air-verse/air@latest

COPY backend/ .

EXPOSE 8080

ENTRYPOINT ["air"]
