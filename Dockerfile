FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /jabline ./cmd/jabline

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /jabline /usr/local/bin/jabline
COPY --from=builder /build/internal/embedded/modules /app/modules
COPY --from=builder /build/registry /app/registry
ENTRYPOINT ["jabline"]
CMD ["--help"]
