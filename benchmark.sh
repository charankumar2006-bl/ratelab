#!/bin/bash

BASE_URL="http://localhost:8080"
TOTAL_REQUESTS=20
CLIENT_ID="benchmark-$(date +%s%N)"
RESULT_FILE=$(mktemp)
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

echo "RateLab Load Experiment"
echo "======================="
echo "Requests: $TOTAL_REQUESTS"
echo "Client: $CLIENT_ID"
echo "Sending concurrent requests..."
echo

for i in $(seq 1 "$TOTAL_REQUESTS"); do
    (
        curl -s -o /dev/null \
            -w "%{http_code}\n" \
            -H "X-Client-ID: $CLIENT_ID" \
            "$BASE_URL/api/data" >> "$RESULT_FILE"
    ) &
done

wait

ACCEPTED=$(grep -c '^200$' "$RESULT_FILE")
REJECTED=$(grep -c '^429$' "$RESULT_FILE")
TOTAL=$(wc -l < "$RESULT_FILE")
OTHER=$((TOTAL - ACCEPTED - REJECTED))

if [ "$TOTAL" -gt 0 ]; then
    REJECTION_RATE=$(awk -v rejected="$REJECTED" -v total="$TOTAL" \
        'BEGIN { printf "%.2f", rejected * 100 / total }')
else
    REJECTION_RATE="0.00"
fi

echo "Results"
echo "-------"
echo "Total responses:  $TOTAL"
echo "Accepted (200):   $ACCEPTED"
echo "Rejected (429):   $REJECTED"
echo "Other responses:  $OTHER"
echo "Rejection rate:   $REJECTION_RATE%"

if [ ! -f results.csv ]; then
    echo "timestamp,algorithm,total,accepted,rejected,other,rejection_rate_percent" > results.csv
fi

echo "$TIMESTAMP,unknown,$TOTAL,$ACCEPTED,$REJECTED,$OTHER,$REJECTION_RATE" >> results.csv

echo
echo "Results saved to: $PWD/results.csv"

rm -f "$RESULT_FILE"