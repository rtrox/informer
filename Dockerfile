FROM golang:1.20-alpine AS build_base
WORKDIR /tmp/informer

ARG VERSION="devel"
ARG BUILDTIME="unknown"
ARG REVISION="unknown"

RUN apk --update add \
    ca-certificates \
    gcc \
    g++

# Copy go.mod and go.sum first to leverage Docker layer cache
COPY go.mod ./
COPY go.sum ./
# Cache dependency packages
RUN go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get

# Copy only source code directories to minimize cache misses
COPY . .

# CGO is required by the sqlite3 driver
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
        CGO_ENABLED=1 go build \
        -ldflags="-linkmode external -extldflags '-static' -s -w -X main.version=${VERSION}" \
        -o /tmp/informer/out/informer \
         ./cmd/informer/main.go

FROM scratch
COPY --from=build_base /tmp/informer/out/informer /bin/informer
COPY --from=build_base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 8080/tcp
ENTRYPOINT ["/bin/informer"]