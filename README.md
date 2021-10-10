# Tracing & Monitoring

## Prometheus Queries

### Histogram

* The average request duration within the last 5 minutes

```
rate(rpc_durations_histogram_seconds_sum[5m])/rate(rpc_durations_histogram_seconds_count[5m])
```

* Calculation by percentile

```
histogram_quantile(0.95, sum(rate(rpc_durations_histogram_seconds_bucket[5m])) by (le))
```

* Apdex score
* If you have an objective to serve 95% of requests within 300ms, we can express the relative amount of requests servered within that upper boundary
* For this example, we will us 12.5s as our upper bound

```
sum(rate(rpc_durations_histogram_seconds_bucket{le="10"}[60m])) / sum(rate(rpc_durations_histogram_seconds_count[60m])) 
```

### Summary

* The average request duration within the last 5 minutes

```
rate(rpc_durations_summary_seconds_sum[5m])/rate(rpc_durations_summary_seconds_count[5m])
```

* Calculation by pre-defined quantile
* Must be exactly as defined in application

```
rpc_durations_summary_seconds{quantile="0.9"}
```