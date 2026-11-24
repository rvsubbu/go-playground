# Go HTTP Server Playground

A small Go project for illustrating HTTP server fundamentals and comparing
`net/http`, Chi, and Gin. The `basic_server` app exposes the same API with each
router; select the implementation in YAML configuration; graceful exit handled;
basic metrics for requests_total and latency added

## Requirements

- Go 1.27 or newer
- Make

## Run

From the repository root, start the server using the configuration in
[`config/config.yaml`](./config/config.yaml):

```sh
make run_server
```

The default server address is `:8080`. To use another configuration file:

```sh
go run ./basic_server -config path/to/config.yaml
```

The config path is resolved relative to the current working directory. Select
`net/http`, `chi`, or `gin` with `server.web_server`:

```yaml
app:
  name: api_server
  env: dev

logging:
  level: debug

server:
  addr: ":8080"
  read_timeout: 1s
  web_server: gin
```

At `debug` level, the server logs request start and completion, including the
method, path, status, and time taken.

## API

All endpoints use `GET`:

| Path | Purpose |
| --- | --- |
| `/version` | Build version, Git revision, and build time |
| `/config` | Active application configuration |
| `/healthz` | Liveness response |
| `/metrics` | Prometheus text-format metrics |

The metrics endpoint exposes `http_requests_total`, labeled by `app_name`,
`method`, `status`, and `apiNickname` (`version`, `config`, `healthz`,
`metrics`, or `unknown`). It also exposes the
`http_request_duration_seconds` latency histogram.

## Build and test

```sh
make build_server
make test
```

`make build_server` writes a platform-specific binary under `bin/`.
`make test` runs the Go tests with the race detector and reports aggregate
coverage. Use `make clean` to remove build artifacts and the test cache.

## Layout

- `basic_server/` — server setup and router selection
- `config/` — default YAML configuration
- `util/` — shared handlers, configuration loading, and metrics
