# syntax=docker/dockerfile:1
# Minimal secure multi-stage build for risinggo MCP server
# Non-root, distroless, small attack surface. Supports k6 loadtest compose.

FROM golang:1.25-bookworm AS builder
WORKDIR /src

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /out/server ./cmd/server

# Runtime: distroless static (no shell, minimal)
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/server /server

# Default to HTTP transport on standard port for k6 / compose
ENV MCP_TRANSPORT=streamable-http \
    MCP_HTTP_PORT=8000 \
    MCP_READ_ONLY=true

USER nonroot:nonroot
EXPOSE 8000

ENTRYPOINT ["/server"]
