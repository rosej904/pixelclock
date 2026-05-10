# ── Stage 1: Build ────────────────────────────────────────────────────────────
# Use the full Go SDK image just for compiling. This image is large but is
# never shipped — it's only used during the build phase.
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first so Docker can cache this layer.
# If go.mod/go.sum don't change, Docker skips re-downloading modules.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source and build the binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o pixelclock ./cmd/clock

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
# Scratch is a completely empty image — no shell, no OS, just our binary.
# The final image ends up ~10MB instead of ~300MB.
FROM scratch

# Copy only the compiled binary from the builder stage.
COPY --from=builder /app/pixelclock /pixelclock

ENTRYPOINT ["/pixelclock"]
