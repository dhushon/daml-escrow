# -----------------------------
# Build Stage
# -----------------------------
FROM golang:1.22-alpine AS builder

WORKDIR /app

# install git for module resolution
RUN apk add --no-cache git

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

COPY --from=builder /app/escrow-api .

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/escrow-api"]