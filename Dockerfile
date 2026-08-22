FROM golang:1.25-bookworm AS backend
WORKDIR /app
COPY backend/ ./backend/
RUN cd backend/byteport && go build -o /byteport-server

FROM rust:1.85-slim AS desktop
WORKDIR /app
COPY . .
RUN cargo build --release -p byteport-cli

FROM debian:bookworm-slim
COPY --from=backend /byteport-server /usr/local/bin/
COPY --from=desktop /app/target/release/byteport-cli /usr/local/bin/
CMD ["byteport-server"]
