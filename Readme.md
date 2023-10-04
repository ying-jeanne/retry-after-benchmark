Server/Client Benchmarking

Within this project, you have the flexibility to launch various servers, each of which exports request total metrics to Prometheus. You can find the relevant Prometheus configuration file in the "prometheus" folder.

Once everything is set up and operational, visit http://localhost:9090/graph. Here, you can conveniently benchmark different retry-after algorithms using the rate(http_requests_total[1s]) metric.
