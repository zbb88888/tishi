# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
        -X github.com/zbb88888/tishi/internal/cmd.Version=${VERSION} \
        -X github.com/zbb88888/tishi/internal/cmd.GitCommit=${COMMIT} \
        -X github.com/zbb88888/tishi/internal/cmd.BuildDate=${DATE}" \
    -o /tishi ./cmd/tishi

# Runtime stage — distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /tishi /tishi

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/tishi"]
CMD ["server"]
