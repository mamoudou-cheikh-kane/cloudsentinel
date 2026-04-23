# syntax=docker/dockerfile:1.6

# -----------------------------------------------------------------------------
# Stage 1: build
# -----------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Install build deps (git for go modules over HTTPS+SSH fallback, ca-certificates).
RUN apk add --no-cache git ca-certificates

# Download deps first (leverages Docker layer cache).
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source.
COPY . .

# Build a static binary. CGO_ENABLED=0 avoids libc dependency.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/agent \
    ./cmd/agent

# -----------------------------------------------------------------------------
# Stage 2: minimal runtime
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/agent /agent

EXPOSE 9100 50051

USER nonroot:nonroot

ENTRYPOINT ["/agent"]
