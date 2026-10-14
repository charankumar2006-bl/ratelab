# RateLab — Experimental Results

## 1. Objective

Compare Fixed Window and Sliding Window rate-limiting algorithms
using a Go API gateway and Redis.

The configured limit is five requests per client per 60 seconds.

## 2. Boundary-Burst Experiment

The experiment sent five requests before a minute boundary and
five requests immediately after it.

| Metric | Fixed Window | Sliding Window |
|---|---:|---:|
| Requests before boundary | 5 | 5 |
| Accepted before boundary | 5 | 5 |
| Requests after boundary | 5 | 5 |
| Accepted after boundary | 5 | 0 |
| Total accepted | 10 | 5 |
| Total rejected | 0 | 5 |

### Observation

Fixed Window can permit a burst across a window boundary.
Sliding Window continues considering recent requests and rejected
the second batch in this experiment.

## 3. Concurrent Request Experiment

Each benchmark sent 100 concurrent requests using one client ID.

| Metric | Fixed Window run | Sliding-labelled run |
|---|---:|---:|
| Total responses | 100 | 100 |
| Accepted | 5 | 5 |
| Rejected | 95 | 95 |
| Rejection rate | 95% | 95% |

The second run was labelled Sliding Window by the benchmark prompt.
The gateway's active algorithm was not independently verified at
the time of that run, so this comparison is preliminary.

## 4. Performance Summary

The analysis script reported the following results:

| Metric | Fixed Window | Sliding Window |
|---|---:|---:|
| Trials | 1 | 1 |
| Throughput (responses/sec) | 434.97 | 398.53 |
| Average latency (ms) | 22.638 | 72.020 |
| Average accepted requests | 5 | 5 |
| Average rejected requests | 95 | 95 |

These are single-trial measurements. They should not be interpreted
as conclusive evidence that one algorithm is generally faster.

Different benchmark scripts produced different measurements.
The scripts and their source CSV files should be checked before
combining results into a single comparison.

## 5. Sustained-Load Experiment

One 30-second run used five clients.

| Metric | Result |
|---|---:|
| Duration | 30 seconds |
| Clients | 5 |
| Total responses | 4,490 |
| Accepted | 25 |
| Rejected | 4,465 |
| Other responses | 0 |
| Rejection rate | 99.44% |
| Response rate | Approximately 149.7 responses/sec |

The 25 accepted requests match the configured allowance of five
requests for each of five clients.

## 6. Resource Measurements

The resource monitor recorded approximately 17.4–17.7 MB RSS
in the displayed samples.

CPU usage was reported as 0.0% and memory usage as 0.1%.
The CPU measurement method needs further validation before these
values can be treated as reliable measurements under load.

## 7. Conclusions

1. Both algorithms enforced the five-request limit in the tested
   concurrent-request scenarios.
2. Fixed Window permitted a boundary burst in the recorded test.
3. Sliding Window rejected the second batch in the corresponding
   boundary-burst test.
4. Performance measurements varied between runs and scripts.
5. More repeated, independently verified trials are needed to
   make a reliable performance comparison.

## 8. Limitations

- Performance trials were limited and not always repeated equally.
- The active algorithm was not independently verified for every
  historical benchmark run.
- Local machine load can affect latency and throughput.
- Resource-monitoring CPU figures require validation.
- Results describe this test setup and do not establish universal
  performance rankings.

## 9. Reproducibility

Project root: `~/Documents/ratelab`

Main technologies:
- Go
- Redis
- Docker Compose
- Bash
- Python

Relevant result files:
- `boundary_results.csv`
- `algorithm_comparison.csv`
- `performance_results_detailed.csv`
- `performance_summary.csv`
- `resource_results.csv`
