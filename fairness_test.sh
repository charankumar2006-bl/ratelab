#!/bin/bash

BASE_URL="http://localhost:8080"
TOTAL_CLIENTS=5
REQUESTS_PER_CLIENT=10
RESULT_FILE=$(mktemp)

trap 'rm -f "$RESULT_FILE"' EXIT

echo "===================================="
echo "RateLab Multi-Client Fairness Test"
echo "===================================="
echo "Clients: $TOTAL_CLIENTS"
echo "Requests per client: $REQUESTS_PER_CLIENT"
echo "Rate limit: 5 requests per 60 seconds"
echo

for client_number in $(seq 1 "$TOTAL_CLIENTS"); do
    (
        client_id="fairness-client-$client_number"

        for request_number in $(seq 1 "$REQUESTS_PER_CLIENT"); do
            status=$(curl -s -o /dev/null \
                --max-time 10 \
                -w "%{http_code}" \
                -H "X-Client-ID: $client_id" \
                "$BASE_URL/api/data")

            printf '%s,%s\n' "$client_id" "$status" >> "$RESULT_FILE"
        done
    ) &
done

wait

echo
echo "Results per client"
echo "------------------"

for client_number in $(seq 1 "$TOTAL_CLIENTS"); do
    client_id="fairness-client-$client_number"

    accepted=$(awk -F, -v id="$client_id" \
        '$1 == id && $2 == 200 {n++} END {print n+0}' "$RESULT_FILE")

    rejected=$(awk -F, -v id="$client_id" \
        '$1 == id && $2 == 429 {n++} END {print n+0}' "$RESULT_FILE")

    other=$(awk -F, -v id="$client_id" \
        '$1 == id && $2 != 200 && $2 != 429 {n++} END {print n+0}' "$RESULT_FILE")

    echo "$client_id: accepted=$accepted rejected=$rejected other=$other"
done

echo
echo "Experiment complete."