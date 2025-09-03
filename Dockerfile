# ---------- build stage ----------
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o alertmanager-relay cmd/alertmanager-relay/main.go

FROM ghcr.io/meck93/distroless-http-healthcheck:latest AS healthcheck

# ---------- runtime stage ----------
FROM alpine:3.22

COPY --from=healthcheck /healthcheck /healthcheck

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /src/alertmanager-relay /app/alertmanager-relay

EXPOSE 8080
ENTRYPOINT ["/app/alertmanager-relay"]
