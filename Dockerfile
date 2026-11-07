# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build
ARG COMMIT=unknown
ARG DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o /out/operator ./cmd/operator

# Build CLI
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/quantumctl ./cmd/quantumctl

# Stage 2: Runtime
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /out/operator /operator
COPY --from=builder /out/quantumctl /quantumctl

USER nonroot:nonroot

ENTRYPOINT ["/operator"]
