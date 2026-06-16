# High-Assurance Tripartite API Dockerfile
# Authoritatively supports Multi-Arch builds (ARM64/AMD64).

# Stage 1: Build
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25-alpine AS builder

ARG TARGETARCH
ARG TARGETOS

WORKDIR /app

# Dependencies
RUN apk add --no-cache git
COPY go.mod go.sum ./
COPY third_party/ ./third_party/
RUN go mod download

# Source and Build
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o escrow-api ./cmd/escrow-api

# Stage 2: Minimal Identity
FROM gcr.io/distroless/base-debian12:latest

WORKDIR /app
COPY --from=builder /app/escrow-api .
COPY --from=builder /app/config ./config
COPY --from=builder /app/architecture ./architecture

EXPOSE 8081

ENTRYPOINT ["/app/escrow-api"]
CMD ["serve"]
