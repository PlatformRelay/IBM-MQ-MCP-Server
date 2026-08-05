# Build the ibm-mq-mcp binary (CGO-free).
FROM golang:1.26.5 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-w -s" -o ibm-mq-mcp ./cmd/ibm-mq-mcp

# Minimal nonroot runtime. HEALTHCHECK and k8s securityContext readOnlyRootFilesystem
# guidance belong in OBS-001 / deployment docs, not this image build.
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/ibm-mq-mcp /ibm-mq-mcp
USER 65532:65532

ENTRYPOINT ["/ibm-mq-mcp"]
