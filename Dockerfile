FROM golang:1.21-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gerdu .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && adduser -D -u 10001 gerdu
COPY --from=builder /out/gerdu /usr/local/bin/gerdu
USER gerdu
EXPOSE 8080 8081 11211 6379 12000
ENTRYPOINT ["/usr/local/bin/gerdu"]
