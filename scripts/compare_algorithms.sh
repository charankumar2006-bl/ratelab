#!/bin/bash

set -u

BASE_URL="http://localhost:8080"
TOTAL_REQUESTS=20
RESULT_FILE="algorithm_comparison_verified.csv"

# Verify the gateway is running and identify its actual algorithm.
health=$(curl -fsS --max-time 3 "$BASE_URL/health") || {
    echo "ERROR: Gateway is not responding at $BASE_URL"
    exit 1
}

algorithm=$(printf '%s' "$health" | python3 -c '
import json
import sys
try:
    print(json.load(sys.stdin)["algorithm"])
except (ValueError, KeyError, TypeError):
    sys.exit(1)
') || {
    echo "ERROR: Could not identify the active algorithm."
    echo "Health response: $health"
    exit 1
}

if [[ "$algorithm" != "fixed" && "$algorithm" != "sliding" ]]; then
    echo "ERROR: Unknown algorithm: $algorithm"
    exit 1
fi

# Verify that the backend is available.
curl -fsS --max-time 3 \
    -o /dev/null \
    -H "X-Client-ID: ratelab-health-check" \
    "$BASE_URL/api/data" || {
    echo "ERROR: API/backend check failed."
    exit 1
}

# Create a new results file with a header if necessary.
if [ ! -s "$RESULT_FILE" ]; then
    echo "algorithm,total,accepted,rejected,other,rejection_rate_percent" \
        > "$RESULT_FILE"
fi

echo
echo "RateLab Algorithm Comparison"
echo "Active algorithm: $algorithm"
echo "Total requests: $TOTAL_REQUESTS"

client_id="compare-${algorithm}-$(date +%s%N)"
temp_file=$(mktemp)
trap 'rm -f "$temp_file"' EXIT

# Send concurrent requests using one client ID.
for i in $(seq 1 "$TOTAL_REQUESTS"); do
    curl -sS --max-time 15 \
        -o /dev/null \
        -w "%{http_code}\n" \
        -H "X-Client-ID: $client_id" \
        "$BASE_URL/api/data" >> "$temp_file" &
done

wait

total=$(wc -l < "$temp_file")
accepted=$(grep -c '^200$' "$temp_file" || true)
rejected=$(grep -c '^429$' "$temp_file" || true)
other=$((total - accepted - rejected))

rejection_rate=$(awk -v r="$rejected" -v t="$total" \
    'BEGIN {
        if (t > 0) printf "%.2f", r * 100 / t
        else print "0.00"
    }')

echo
echo "Results"
echo "Algorithm: $algorithm"
echo "Total responses: $total"
echo "Accepted: $accepted"
echo "Rejected: $rejected"
echo "Other responses: $other"
echo "Rejection rate: $rejection_rate%"

if [ "$total" -ne "$TOTAL_REQUESTS" ]; then
    echo "WARNING: Not all requests produced an HTTP response."
    echo "Inspect the gateway and backend before trusting this run."
fi

echo "$algorithm,$total,$accepted,$rejected,$other,$rejection_rate" \
    >> "$RESULT_FILE"

echo
echo "Results saved to $RESULT_FILE"
