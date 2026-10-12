import csv
import matplotlib.pyplot as plt

input_file = "boundary_results.csv"
output_file = "boundary_comparison.png"

algorithms = []
accepted = []
rejected = []

with open(input_file, newline="") as file:
    reader = csv.DictReader(file)

    for row in reader:
        algorithms.append(row["algorithm"].title())
        accepted.append(int(row["total_accepted"]))
        rejected.append(int(row["total_rejected"]))

positions = list(range(len(algorithms)))
width = 0.35

plt.figure(figsize=(9, 5))

plt.bar(
    [p - width / 2 for p in positions],
    accepted,
    width,
    label="Accepted requests",
)

plt.bar(
    [p + width / 2 for p in positions],
    rejected,
    width,
    label="Rejected requests",
)

plt.xticks(positions, algorithms)
plt.ylabel("Number of requests")
plt.xlabel("Rate-limiting algorithm")
plt.title("RateLab: Boundary-Burst Experiment")
plt.legend()
plt.tight_layout()

plt.savefig(output_file, dpi=200)
print(f"Graph saved to {output_file}")

plt.show()
