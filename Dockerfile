# High-Assurance Tripartite API Dockerfile
# Authoritatively supports Multi-Arch builds (ARM64/AMD64).

# Stage 1: High-Assurance Build
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25-alpine AS builder

ARG TARGETARCH
ARG TARGETOS

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Pre-fetch dependencies (Optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o escrow-api ./cmd/escrow-api

# Stage 2: Minimal Production Identity
FROM gcr.io/distroless/base-debian12:latest

WORKDIR /app

# Authoritatively copy the architecture-optimized binary
COPY --from=builder /app/escrow-api .
COPY --from=builder /app/config ./config

# High-Assurance Defaults
EXPOSE 8081

ENTRYPOINT ["/app/escrow-api"]
CMD ["serve"]
