FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app -h /app app

WORKDIR /app

COPY --from=builder /out/server ./server
COPY --from=builder /out/migrate ./migrate
COPY docs/openapi.yaml ./docs/openapi.yaml
COPY migrations ./migrations

RUN chmod 0555 ./server ./migrate \
    && chown -R app:app /app

USER app

EXPOSE 8080

CMD ["./server"]
