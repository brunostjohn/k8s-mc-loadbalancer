FROM golang:1.22-bookworm as builder
  WORKDIR /app
  COPY ./ ./
  RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x
  ENV GOCACHE=/root/.cache/go-build
  RUN --mount=type=cache,target="/root/.cache/go-build" go build -v -o server

FROM alpine:3.12
  RUN addgroup --gid 1001 appuser \
   && adduser --uid 1001 -G appuser -h /app -D appuser \
   && chown -R 1001:1001 /app
  USER appuser
  WORKDIR /app
  COPY --from=builder --chown=1001:1001 /app/server /app/server
  CMD ["/app/server"]
