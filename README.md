# alertmanager-relay

Simple service to relay alerts from one alertmanager to another.

## Usage

Run with Docker.

```
docker run --rm -e SRC_ALERTMANAGER_URL=http://source:9093 \
                 -e DST_ALERTMANAGER_URL=http://dest:9093 \
                 -e DST_AUTH_USERNAME=alice \
                 -e DST_AUTH_PASSWORD=s3cr3t \
                 -p 8080:8080 \
                 alertmanager-relay:latest
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

## TODO

- kustomize
