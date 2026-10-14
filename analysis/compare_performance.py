import csv
from collections import defaultdict
from statistics import mean

INPUT_FILE = "performance_results_detailed.csv"
OUTPUT_FILE = "performance_summary.csv"

metrics = {
    "throughput_responses_per_second": "Average throughput (responses/sec)",
    "mean_latency_ms": "Average latency (ms)",
    "p50_latency_ms": "P50 latency (ms)",
    "p95_latency_ms": "P95 latency (ms)",
    "rejection_rate_percent": "Rejection rate (%)",
}

results = defaultdict(list)

try:
    with open(INPUT_FILE, "r", newline="") as file:
        reader = csv.DictReader(file)

        for row in reader:
            algorithm = row["algorithm"].strip().lower()

            if algorithm not in ("fixed", "sliding"):
                continue

            results[algorithm].append(row)

except FileNotFoundError:
    raise SystemExit(
        f"{INPUT_FILE} not found. Run performance_test.py first."
    )

if not results:
    raise SystemExit("No valid benchmark results found.")

summary_rows = []

print("=" * 70)
print("              RATELAB PERFORMANCE COMPARISON")
print("=" * 70)

for algorithm in ("fixed", "sliding"):
    trials = results.get(algorithm, [])

    if not trials:
        print(f"\n{algorithm}: no trials recorded")
        continue

    print(f"\nAlgorithm: {algorithm.upper()}")
    print(f"Number of trials: {len(trials)}")

    summary = {
        "algorithm": algorithm,
        "trials": len(trials),
    }

    for metric, label in metrics.items():
        values = []

        for trial in trials:
            try:
                value = float(trial[metric])
                values.append(value)
            except (ValueError, KeyError):
                continue

        if values:
            average = mean(values)
            summary[f"average_{metric}"] = round(average, 3)
            print(f"{label}: {average:.3f}")
        else:
            summary[f"average_{metric}"] = ""
            print(f"{label}: unavailable")

    summary_rows.append(summary)

if summary_rows:
    columns = ["algorithm", "trials"] + [
        f"average_{metric}" for metric in metrics
    ]

    with open(OUTPUT_FILE, "w", newline="") as file:
        writer = csv.DictWriter(file, fieldnames=columns)
        writer.writeheader()
        writer.writerows(summary_rows)

    print(f"\nSummary saved to {OUTPUT_FILE}")