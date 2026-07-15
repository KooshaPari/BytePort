# =============================================================================
# BytePort Rust CLI — multi-stage Dockerfile
#
# Build targets (use --target):
#   cli       — byteport CLI binary (default)
#   full      — CLI + entrypoint shell
# =============================================================================
# Stage 1: Rust build environment
FROM rust:1.85-alpine AS chef

RUN apk add --no-cache musl-dev pkg-config openssl-dev

# Install cargo-chef for layer caching
RUN cargo install cargo-chef --locked

WORKDIR /build

FROM chef AS planner
COPY . .
RUN cargo chef prepare --recipe-path recipe.json

FROM chef AS builder
COPY --from=planner /build/recipe.json recipe.json
RUN cargo chef cook --release --recipe-path recipe.json

COPY . .

# Build the CLI binary (uses the workspace root)
RUN cargo build --release -p byteport-cli && \
    cp target/release/byteport /out/byteport

# =============================================================================
# Stage 2: Minimal runtime image
FROM alpine:3.21 AS cli

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/byteport /usr/local/bin/byteport

ENTRYPOINT ["/usr/local/bin/byteport"]

# =============================================================================
# Stage 3: Full image with shell (useful for debugging)
FROM alpine:3.21 AS full

RUN apk add --no-cache ca-certificates tzdata bash curl jq

COPY --from=builder /out/byteport /usr/local/bin/byteport

CMD ["/usr/local/bin/byteport"]
