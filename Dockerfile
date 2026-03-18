# -----------------------------
# Build Stage
# -----------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# install git for module resolution
RUN apk add --no-cache git

# copy local dependencies needed for go mod download
COPY third_party ./third_party

# cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy source
COPY . .

# build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -o escrow-api ./cmd/escrow-api


# -----------------------------
# Runtime Stage
# -----------------------------
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy binary and config
COPY --from=builder /app/escrow-api .
COPY --from=builder /app/config ./config

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/escrow-api"]
