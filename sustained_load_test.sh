#!/bin/bash

BASE_URL="http://localhost:8080"
DURATION=30
CLIENTS=5
RESULT_FILE=$(mktemp)

trap 'rm -f "$RESULT_FILE"' EXIT

echo "===================================="
echo "RateLab Sustained-Load Experiment"
echo "===================================="
echo "Duration: $DURATION seconds"
echo "Clients: $CLIENTS"
echo "Rate limit: 5 requests per client per 60 seconds"
echo

START=$(date +%s)

for client_number in $(seq 1 "$CLIENTS"); do
    (
        client_id="sustained-client-$client_number"

        while [ "$(( $(date +%s) - START ))" -lt "$DURATION" ]; do
            curl -s -o /dev/null \
                --max-time 3 \
                -w '%{http_code}\n' \
                -H "X-Client-ID: $client_id" \
                "$BASE_URL/api/data" >> "$RESULT_FILE"
        done
    ) &
done

wait

TOTAL=$(wc -l < "$RESULT_FILE")
ACCEPTED=$(grep -c '^200$' "$RESULT_FILE" || true)
REJECTED=$(grep -c '^429$' "$RESULT_FILE" || true)
OTHER=$((TOTAL - ACCEPTED - REJECTED))

echo
echo "Results"
echo "------------------------------------"
echo "Total responses: $TOTAL"
echo "Accepted (200):  $ACCEPTED"
echo "Rejected (429):  $REJECTED"
echo "Other responses: $OTHER"
echo "------------------------------------"