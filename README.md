# service-telemetry
An example web-service which exposes prometheus compatible metrics

## Resources

### /hello/{name}

Returns `Hello {name}` with `HTTP 200 OK` status. 
Use **/hello/{number}** to simulate a long running task resulting in high latency.

### /healthz

Returns `HTTP 200 OK` status. 
This resource will be used by Kubelet health checks.

### /metrics

Returns Prometheus compatible metrics. 
This is a standard Prometheus handler.