# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/server ./cmd/server \
 && apk add --no-cache upx \
 && upx --best --lzma /out/server

FROM scratch
WORKDIR /app
COPY --from=builder /out/server /app/server
COPY --from=builder /src/config/db/migrations /app/config/db/migrations
USER 65532:65532
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=4s --start-period=15s --retries=3 \
  CMD ["/app/server", "healthcheck"]
ENTRYPOINT ["/app/server"]
