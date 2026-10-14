#!/bin/bash

BASE_URL="http://localhost:8080"
CLIENT_ID="boundary-$(date +%s%N)"

echo "=================================="
echo "RateLab Boundary-Burst Experiment"
echo "=================================="
echo "Client ID: $CLIENT_ID"
echo "Limit: 5 requests per 60 seconds"
echo

# Calculate the next 60-second window boundary.
NOW=$(date +%s)
BOUNDARY=$(( (NOW / 60 + 1) * 60 ))
WAIT=$(( BOUNDARY - NOW - 2 ))

# Ensure there is enough time to send the first batch.
if [ "$WAIT" -lt 1 ]; then
    BOUNDARY=$(( BOUNDARY + 60 ))
    WAIT=$(( BOUNDARY - NOW - 2 ))
fi

echo "Waiting until 2 seconds before the window boundary..."
sleep "$WAIT"

echo
echo "=== Batch 1: Before boundary ==="

for i in {1..5}; do
    STATUS=$(curl -s -o /dev/null \
        -w "%{http_code}" \
        --max-time 5 \
        -H "X-Client-ID: $CLIENT_ID" \
        "$BASE_URL/api/data")

    echo "Request $i: HTTP ${STATUS:-FAILED}"
done

# Wait until one second after the boundary.
NOW=$(date +%s)
WAIT=$(( BOUNDARY + 1 - NOW ))

if [ "$WAIT" -gt 0 ]; then
    sleep "$WAIT"
fi

echo
echo "=== Batch 2: After boundary ==="

for i in {6..10}; do
    STATUS=$(curl -s -o /dev/null \
        -w "%{http_code}" \
        --max-time 5 \
        -H "X-Client-ID: $CLIENT_ID" \
        "$BASE_URL/api/data")

    echo "Request $i: HTTP ${STATUS:-FAILED}"
done

echo
echo "Experiment complete."