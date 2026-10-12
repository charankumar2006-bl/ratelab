import csv
from collections import defaultdict
from statistics import mean

filename = "performance_results.csv"
results = defaultdict(list)

try:
    with open(filename, "r", newline="") as file:
        reader = csv.DictReader(file)

        for row in reader:
            algorithm = row["algorithm"]

            results[algorithm].append({
                "throughput": float(
                    row["throughput_responses_per_second"]
                ),
                "latency": float(row["average_latency_ms"]),
                "accepted": int(row["accepted"]),
                "rejected": int(row["rejected"]),
                "total": int(row["total"]),
            })

except FileNotFoundError:
    raise SystemExit(
        f"{filename} was not found. Run performance_test.sh first."
    )

if not results:
    raise SystemExit("No benchmark results found in the CSV file.")

print("=" * 65)
print("              RATELAB PERFORMANCE ANALYSIS")
print("=" * 65)

for algorithm, trials in sorted(results.items()):
    avg_throughput = mean(t["throughput"] for t in trials)
    avg_latency = mean(t["latency"] for t in trials)
    avg_accepted = mean(t["accepted"] for t in trials)
    avg_rejected = mean(t["rejected"] for t in trials)

    print(f"\nAlgorithm: {algorithm}")
    print(f"Trials: {len(trials)}")
    print(f"Average throughput: {avg_throughput:.2f} responses/second")
    print(f"Average latency: {avg_latency:.3f} ms")
    print(f"Average accepted requests: {avg_accepted:.1f}")
    print(f"Average rejected requests: {avg_rejected:.1f}")

print("\n" + "=" * 65)
print("Analysis complete.")