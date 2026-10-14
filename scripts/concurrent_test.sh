#!/bin/bash

BASE_URL="http://localhost:8080"
CLIENT_ID="concurrent-test-$(date +%s%N)"
TOTAL_REQUESTS=20

RESULT_FILE=$(mktemp)
trap 'rm -f "$RESULT_FILE"' EXIT

echo "========================================"
echo " RateLab Concurrent Traffic Experiment"
echo "========================================"
echo "Client: $CLIENT_ID"
echo "Total requests: $TOTAL_REQUESTS"
echo "Configured limit: 5 requests per window"
echo "----------------------------------------"

for i in $(seq 1 "$TOTAL_REQUESTS"); do
    curl -s -o /dev/null \
        -w "%{http_code}\n" \
        -H "X-Client-ID: $CLIENT_ID" \
        "$BASE_URL/api/data" >> "$RESULT_FILE" &
done

wait

SUCCESS=$(grep -c '^200$' "$RESULT_FILE" || true)
REJECTED=$(grep -c '^429$' "$RESULT_FILE" || true)
OTHER=$(grep -vcE '^(200|429)$' "$RESULT_FILE" || true)

echo
echo "============= RESULTS ============="
echo "Total requests:    $TOTAL_REQUESTS"
echo "Accepted (200):    $SUCCESS"
echo "Rejected (429):    $REJECTED"
echo "Other responses:   $OTHER"
echo "==================================="
