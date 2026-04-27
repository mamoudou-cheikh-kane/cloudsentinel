# syntax=docker/dockerfile:1.6
# -----------------------------------------------------------------------------
# Stage 1: build the Go binary
# -----------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder
WORKDIR /src

# Build deps: git for go modules over HTTPS+SSH fallback,
# ca-certificates for TLS verification at module download time.
RUN apk add --no-cache git ca-certificates

# Download deps first (leverages Docker layer cache).
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source.
COPY . .

# Build a static binary. CGO_ENABLED=0 avoids any libc dependency,
# -s -w strips the binary, -trimpath removes local paths so the
# image stays reproducible across build hosts.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/agent \
    ./cmd/agent

# -----------------------------------------------------------------------------
# Stage 2: collect tc and the shared libraries it needs
# -----------------------------------------------------------------------------
# We pull tc (from iproute2) plus its runtime dependencies from a
# Debian slim image so the libraries are glibc-based and compatible
# with distroless/base in stage 3. Doing this in a dedicated stage
# keeps the runtime image small: we copy only what we need, not the
# entire apt cache.
FROM debian:bookworm-slim AS tc-stage
RUN apt-get update \
    && apt-get install -y --no-install-recommends iproute2 \
    && rm -rf /var/lib/apt/lists/*

# -----------------------------------------------------------------------------
# Stage 3: minimal runtime
# -----------------------------------------------------------------------------
# distroless/base provides glibc and a non-root user (UID/GID 65532),
# which is the minimum we need for the agent + tc to run safely.
FROM gcr.io/distroless/base-debian12:nonroot

# The agent binary itself.
COPY --from=builder /out/agent /agent

# tc and its dynamic dependencies. The exact set is what `ldd /usr/sbin/tc`
# reports on Debian bookworm; if a library is missing at runtime, the
# pod will crash-loop with a clear "error while loading shared
# libraries" message and we can add it here.
COPY --from=tc-stage /usr/sbin/tc /usr/sbin/tc
COPY --from=tc-stage /usr/lib/x86_64-linux-gnu/libbpf.so.1 /usr/lib/x86_64-linux-gnu/
COPY --from=tc-stage /usr/lib/x86_64-linux-gnu/libelf.so.1 /usr/lib/x86_64-linux-gnu/
COPY --from=tc-stage /usr/lib/x86_64-linux-gnu/libmnl.so.0 /usr/lib/x86_64-linux-gnu/
COPY --from=tc-stage /usr/lib/x86_64-linux-gnu/libxtables.so.12 /usr/lib/x86_64-linux-gnu/
COPY --from=tc-stage /usr/lib/x86_64-linux-gnu/libz.so.1 /usr/lib/x86_64-linux-gnu/
COPY --from=tc-stage /usr/lib/x86_64-linux-gnu/libzstd.so.1 /usr/lib/x86_64-linux-gnu/
COPY --from=tc-stage /usr/lib/x86_64-linux-gnu/libcap.so.2 /usr/lib/x86_64-linux-gnu/
COPY --from=tc-stage /usr/lib/x86_64-linux-gnu/libbsd.so.0 /usr/lib/x86_64-linux-gnu/
COPY --from=tc-stage /usr/lib/x86_64-linux-gnu/libmd.so.0 /usr/lib/x86_64-linux-gnu/
COPY --from=tc-stage /lib/x86_64-linux-gnu/libm.so.6 /usr/lib/x86_64-linux-gnu/
EXPOSE 9100 50051
USER nonroot:nonroot
ENTRYPOINT ["/agent"]
