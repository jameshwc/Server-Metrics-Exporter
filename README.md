# Server Metrics Exporter (for Prometheus)

Prometheus exporter for general server metrics (including CPU, memory, and disk usage)

## Running (Docker)

```bash
$ export METRICS_PORT=7000
$ docker run -d -p $METRICS_PORT:80 jameshwc/server-metrics-exporter # localhost:$METRICS_PORT/metrics
```

## Running (Build locally)

```bash
$ git clone https://github.com/jameshwc/Server-Metrics-Exporter
$ cd Server-Metrics-Exporter
$ go build -o app # Please change the port number if needed before build
$ ./app 
```

