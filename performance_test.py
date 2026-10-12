import csv
import os
import time
import uuid
import urllib.request
import urllib.error

from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime
from statistics import mean

BASE_URL = "http://localhost:8080/api/data"
HEALTH_URL = "http://localhost:8080/health"

TOTAL_REQUESTS = 100
CONCURRENT_WORKERS = 20

RESULT_FILE = "performance_results_detailed.csv"


def check_gateway():
    """Verify that the gateway is reachable."""
    try:
        with urllib.request.urlopen(HEALTH_URL, timeout=3) as response:
            return response.status == 200
    except Exception:
        return False


def send_request(client_id):
    """Send one request and measure its response time."""
    request = urllib.request.Request(
        BASE_URL,
        headers={"X-Client-ID": client_id},
    )

    start = time.perf_counter()

    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            status = response.status
            response.read()

    except urllib.error.HTTPError as error:
        status = error.code
        error.read()
    except Exception:
        status = 0

    latency_ms = (time.perf_counter() - start) * 1000

    return status, latency_ms


def percentile(values, percentage):
    """Calculate a nearest-rank percentile."""
    if not values:
        return 0.0

    ordered = sorted(values)
    rank = max(1, int((percentage / 100) * len(ordered) + 0.999999))

    return ordered[rank - 1]


def main():
    print("=" * 55)
    print("RateLab Detailed Performance Benchmark")
    print("=" * 55)

    algorithm = input(
        "Which algorithm is running (fixed/sliding)? "
    ).strip().lower()

    if algorithm not in ("fixed", "sliding"):
        print("Invalid algorithm. Enter fixed or sliding.")
        return

    if not check_gateway():
        print("Gateway is unreachable. Start the gateway first.")
        return

    client_id = f"benchmark-{uuid.uuid4().hex}"

    print(f"Algorithm: {algorithm}")
    print(f"Requests: {TOTAL_REQUESTS}")
    print(f"Concurrent workers: {CONCURRENT_WORKERS}")
    print("Sending requests...")

    results = []

    start_time = time.perf_counter()

    with ThreadPoolExecutor(max_workers=CONCURRENT_WORKERS) as executor:
        futures = [
            executor.submit(send_request, client_id)
            for _ in range(TOTAL_REQUESTS)
        ]

        for future in as_completed(futures):
            results.append(future.result())

    elapsed = time.perf_counter() - start_time

    accepted = [latency for status, latency in results if status == 200]
    rejected = [latency for status, latency in results if status == 429]
    other = [latency for status, latency in results if status not in (200, 429)]

    total = len(results)
    accepted_count = len(accepted)
    rejected_count = len(rejected)
    other_count = len(other)

    all_latencies = [latency for _, latency in results]

    throughput = total / elapsed if elapsed > 0 else 0
    rejection_rate = rejected_count / total * 100 if total else 0

    print("\nResults")
    print("-" * 55)
    print(f"Responses:              {total}")
    print(f"Accepted (200):         {accepted_count}")
    print(f"Rejected (429):         {rejected_count}")
    print(f"Other responses/errors: {other_count}")
    print(f"Rejection rate:         {rejection_rate:.2f}%")
    print(f"Elapsed time:           {elapsed:.4f} seconds")
    print(f"Throughput:             {throughput:.2f} responses/second")

    print("\nLatency — all responses")
    print(f"Mean:  {mean(all_latencies):.3f} ms")
    print(f"P50:   {percentile(all_latencies, 50):.3f} ms")
    print(f"P95:   {percentile(all_latencies, 95):.3f} ms")

    if accepted:
        print("\nLatency — successful responses (HTTP 200)")
        print(f"Mean:  {mean(accepted):.3f} ms")
        print(f"P50:   {percentile(accepted, 50):.3f} ms")
        print(f"P95:   {percentile(accepted, 95):.3f} ms")

    if rejected:
        print("\nLatency — rate-limited responses (HTTP 429)")
        print(f"Mean:  {mean(rejected):.3f} ms")
        print(f"P50:   {percentile(rejected, 50):.3f} ms")
        print(f"P95:   {percentile(rejected, 95):.3f} ms")

    columns = [
        "timestamp",
        "algorithm",
        "total",
        "accepted",
        "rejected",
        "other",
        "rejection_rate_percent",
        "elapsed_seconds",
        "throughput_responses_per_second",
        "mean_latency_ms",
        "p50_latency_ms",
        "p95_latency_ms",
        "successful_mean_latency_ms",
        "successful_p95_latency_ms",
        "rejected_mean_latency_ms",
        "rejected_p95_latency_ms",
    ]

    row = {
        "timestamp": datetime.now().isoformat(timespec="seconds"),
        "algorithm": algorithm,
        "total": total,
        "accepted": accepted_count,
        "rejected": rejected_count,
        "other": other_count,
        "rejection_rate_percent": round(rejection_rate, 2),
        "elapsed_seconds": round(elapsed, 6),
        "throughput_responses_per_second": round(throughput, 2),
        "mean_latency_ms": round(mean(all_latencies), 3),
        "p50_latency_ms": round(percentile(all_latencies, 50), 3),
        "p95_latency_ms": round(percentile(all_latencies, 95), 3),
        "successful_mean_latency_ms": round(mean(accepted), 3) if accepted else "",
        "successful_p95_latency_ms": round(percentile(accepted, 95), 3) if accepted else "",
        "rejected_mean_latency_ms": round(mean(rejected), 3) if rejected else "",
        "rejected_p95_latency_ms": round(percentile(rejected, 95), 3) if rejected else "",
    }

    file_exists = os.path.exists(RESULT_FILE)

    with open(RESULT_FILE, "a", newline="") as file:
        writer = csv.DictWriter(file, fieldnames=columns)

        if not file_exists or os.path.getsize(RESULT_FILE) == 0:
            writer.writeheader()

        writer.writerow(row)

    print(f"\nResults saved to {RESULT_FILE}")


if __name__ == "__main__":
    main()