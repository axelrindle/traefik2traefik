# Traefik to Traefik Discovery

> Discover configuration from other Traefik instances

## Configuration

Have a look in the `example/` directories for example configuration.

String values may reference environment variables as `${VAR}`. Use this for credentials, for example `basicAuth: ${TRAEFIK_BASIC_AUTH}`.

## Deployment

The program serves the generated configuration at `GET /traefik` and its readiness at `GET /status`. An upstream Traefik reads the first one through its [HTTP provider](https://doc.traefik.io/traefik/providers/http/):

```yaml
providers:
  http:
    endpoint: http://traefik2traefik:1337/traefik
    pollInterval: 30s
```

`/status` and `/traefik` return `503` until the first sync with the source Traefik succeeded, and again after every failed sync.

The image is `ghcr.io/axelrindle/traefik2traefik`, built for `linux/amd64` and `linux/arm64`. It is based on distroless and runs as user `1000`. There is no shell in it.

### Requirements

- The container must reach the source Traefik API (`traefik.endpoint`) and every host in `generator.targets`.
- If `traefik.endpoint` uses a host name, the program resolves it and sends one ICMP echo to each returned address until one answers. The first one that answers is used for the connection. This works without root through unprivileged ICMP sockets, which need the sysctl `net.ipv4.ping_group_range` to include the user's group. Docker sets it to `0 2147483647` by default. Kubernetes does not, see below. If the log shows `failed to fetch Traefik routers` with a permission error, this setting is the first thing to check.
- Use an IP address as endpoint to skip the resolving step.

### Docker

`docker-compose.yml`:

```yaml
services:
  traefik2traefik:
    image: ghcr.io/axelrindle/traefik2traefik:latest
    restart: unless-stopped
    command:
      - -config
      - /etc/traefik2traefik/config.yml
    environment:
      TRAEFIK_BASIC_AUTH: traefik:traefik
    volumes:
      - ./config.yml:/etc/traefik2traefik/config.yml:ro
    ports:
      - "1337:1337"
    healthcheck:
      test:
        - CMD
        - /usr/local/bin/traefik2traefik
        - -config
        - /etc/traefik2traefik/config.yml
        - -healthcheck
      interval: 30s
      timeout: 5s
      retries: 3
    # only needed if the host does not set it for containers already
    sysctls:
      net.ipv4.ping_group_range: "0 2147483647"
```

`config.yml`:

```yaml
server:
  address: ":1337"

traefik:
  endpoint: https://traefik-api.homelab.local
  basicAuth: ${TRAEFIK_BASIC_AUTH}

generator:
  targets:
    - https://10.10.10.1
    - https://10.10.10.2
  entrypoints:
    - websecure
  passHostHeader: true
```

### Kubernetes

The example keeps everything in one namespace. The config goes into a ConfigMap, the credentials into a Secret. One replica is enough, since the state is rebuilt from the source Traefik on every refresh.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: traefik2traefik
data:
  config.yml: |
    server:
      address: ":1337"
    traefik:
      endpoint: https://traefik-api.homelab.local
      basicAuth: ${TRAEFIK_BASIC_AUTH}
    generator:
      targets:
        - https://10.10.10.1
        - https://10.10.10.2
      entrypoints:
        - websecure
      passHostHeader: true
---
apiVersion: v1
kind: Secret
metadata:
  name: traefik2traefik
stringData:
  TRAEFIK_BASIC_AUTH: traefik:traefik
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: traefik2traefik
spec:
  replicas: 1
  selector:
    matchLabels:
      app: traefik2traefik
  template:
    metadata:
      labels:
        app: traefik2traefik
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        sysctls:
          - name: net.ipv4.ping_group_range
            value: "0 2147483647"
      containers:
        - name: traefik2traefik
          image: ghcr.io/axelrindle/traefik2traefik:latest
          args:
            - -config
            - /etc/traefik2traefik/config.yml
          envFrom:
            - secretRef:
                name: traefik2traefik
          ports:
            - name: http
              containerPort: 1337
          readinessProbe:
            httpGet:
              path: /status
              port: http
          livenessProbe:
            tcpSocket:
              port: http
          securityContext:
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
          volumeMounts:
            - name: config
              mountPath: /etc/traefik2traefik
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: traefik2traefik
---
apiVersion: v1
kind: Service
metadata:
  name: traefik2traefik
spec:
  selector:
    app: traefik2traefik
  ports:
    - name: http
      port: 1337
      targetPort: http
```

Notes:

- The liveness probe only checks the socket. `/status` reports `503` whenever the source Traefik is unreachable. Using it for liveness would restart a healthy pod during an outage of the source.
- The readiness probe uses `/status`, so the Service has no endpoints until the first sync worked.
- `net.ipv4.ping_group_range` is on the list of safe sysctls since Kubernetes 1.18. A cluster with a restrictive Pod Security admission level may still reject it. Use an IP address as `traefik.endpoint` in that case and remove the sysctl.
- Pin a version tag (`0.2.0`) instead of `latest` for anything beyond a first test. Tags are created for every release.
- Point the HTTP provider of the upstream Traefik at `http://traefik2traefik.<namespace>.svc:1337/traefik`.

## License

[MIT](LICENSE)
