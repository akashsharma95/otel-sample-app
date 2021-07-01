# OpenTelemetry Collector Sample App

This sample app presents the typical flow of observability data with multiple
OpenTelemetry Collectors deployed:

- Applications send data directly to a Collector configured to use fewer
 resources, aka the _agent_;
- The agent then forwards the data to Collector(s) that receive data from
 multiple agents. Collectors on this layer typically are allowed to use more
 resources and queue more data;
- The Collector then sends the data to the appropriate backend, in this demo
 Jaeger, Zipkin, and Prometheus;

```shell
docker-compose up -d
```

The demo exposes the following backends:

- Jaeger at http://0.0.0.0:16686
- Zipkin at http://0.0.0.0:9411
- Prometheus at http://0.0.0.0:9090
