FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /usr/local/bin/gorphan ./cmd/gorphan

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /usr/local/bin/gorphan /usr/local/bin/gorphan
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
