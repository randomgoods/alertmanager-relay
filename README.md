# alertmanager-relay

[![build](https://github.com/randomgoods/alertmanager-relay/actions/workflows/build.yml/badge.svg)](https://github.com/randomgoods/alertmanager-relay/actions/workflows/build.yml)

Simple service to relay alerts from one alertmanager to another.

## Usage

Run with Docker.

```
docker run --rm -e SRC_ALERTMANAGER_URL=http://source:9093 \
                 -e DST_ALERTMANAGER_URL=http://dest:9093 \
                 -e DST_AUTH_USERNAME=alice \
                 -e DST_AUTH_PASSWORD=s3cr3t \
                 -p 8080:8080 \
                 randomgoods/alertmanager-relay:latest
```

## Deployment

### Executable

Copy the execuable to appropriate host.

### Docker

Either use `docker-compose.{dev,auth}.yml` as a template for deployment with docker compose
or run the container directly with `docker run` like mentioned above.

### k8s

Deploy with kustomize.

```
kubectl apply -k overlays/production/
```

## Development

### Requirements

- golang installation
- docker.io
- docker-compose-v2
- docker-buildx

### Testing

Create an encrypted password for `config/alertmanager/web.yml`.

```
htpasswd -bnBC 12 "" "s3cr3t" | cut -d ':' -f 2
```

Start test setup.

```
docker compose -f docker-compose.dev.yml up -d
```

Sending test alerts.

```
# curl -H "Content-Type: application/json" -d '[{"labels":{"alertname":"myalert"}}]' localhost:9093/api/v1/alerts # /api/v2
```

Checking for alerts at destination.

```
# curl localhost:9095/api/v1/alerts
```

## License

MIT (c) 2025 randomgoods and contributors
