import csv

filename = "boundary_results.csv"

with open(filename, "r", newline="") as file:
    reader = csv.DictReader(file)
    results = list(reader)

if not results:
    raise SystemExit("No experiment results found.")

print("=" * 55)
print("       RATELAB: BOUNDARY-BURST ANALYSIS")
print("=" * 55)

for result in results:
    algorithm = result["algorithm"]
    total = int(result["requests_before"]) + int(result["requests_after"])
    accepted = int(result["total_accepted"])
    rejected = int(result["total_rejected"])

    acceptance_rate = accepted / total * 100 if total else 0
    rejection_rate = rejected / total * 100 if total else 0

    print(f"\nAlgorithm: {algorithm}")
    print(f"Total requests: {total}")
    print(f"Accepted: {accepted}")
    print(f"Rejected: {rejected}")
    print(f"Acceptance rate: {acceptance_rate:.1f}%")
    print(f"Rejection rate: {rejection_rate:.1f}%")

fixed = next(
    (r for r in results if r["algorithm"] == "fixed"),
    None,
)
sliding = next(
    (r for r in results if r["algorithm"] == "sliding"),
    None,
)

if fixed and sliding:
    fixed_accepted = int(fixed["total_accepted"])
    sliding_accepted = int(sliding["total_accepted"])

    print("\n" + "=" * 55)
    print("COMPARISON")
    print("=" * 55)

    print(f"Fixed Window accepted:   {fixed_accepted}")
    print(f"Sliding Window accepted: {sliding_accepted}")
    print(f"Difference: {fixed_accepted - sliding_accepted} requests")

    print(
        "\nInterpretation: Fixed Window can accept more requests "
        "across adjacent window boundaries. Sliding Window "
        "continues counting recent requests."
    )