#!/bin/bash

BASE_URL="http://localhost:8080"
TOTAL_REQUESTS=100
RESULT_FILE="performance_results.csv"
CLIENT_ID="perf-$(date +%s%N)"

echo "RateLab Performance Benchmark"
echo "============================="

read -r -p "Which algorithm is running (fixed/sliding)? " ALGORITHM

if [[ "$ALGORITHM" != "fixed" && "$ALGORITHM" != "sliding" ]]; then
    echo "Invalid algorithm. Enter fixed or sliding."
    exit 1
fi

if ! curl -fsS --max-time 3 \
    http://localhost:8080/health >/dev/null; then
    echo "Gateway is not reachable. Start it before running the benchmark."
    exit 1
fi

if [ ! -f "$RESULT_FILE" ]; then
    echo "timestamp,algorithm,total,accepted,rejected,other,elapsed_seconds,throughput_responses_per_second,average_latency_ms" \
        > "$RESULT_FILE"
fi

TEMP_FILE=$(mktemp)
trap 'rm -f "$TEMP_FILE"' EXIT

echo "Algorithm: $ALGORITHM"
echo "Requests: $TOTAL_REQUESTS"
echo "Sending concurrent requests..."

START_TIME=$(date +%s%N)

for i in $(seq 1 "$TOTAL_REQUESTS"); do
    curl -s -o /dev/null \
        --max-time 10 \
        -w '%{http_code},%{time_total}\n' \
        -H "X-Client-ID: $CLIENT_ID" \
        "$BASE_URL/api/data" >> "$TEMP_FILE" &
done

wait

END_TIME=$(date +%s%N)

ELAPSED=$(awk -v s="$START_TIME" -v e="$END_TIME" \
    'BEGIN { printf "%.6f", (e-s)/1000000000 }')

TOTAL=$(wc -l < "$TEMP_FILE")
ACCEPTED=$(awk -F, '$1 == 200 {n++} END {print n+0}' "$TEMP_FILE")
REJECTED=$(awk -F, '$1 == 429 {n++} END {print n+0}' "$TEMP_FILE")
OTHER=$((TOTAL - ACCEPTED - REJECTED))

THROUGHPUT=$(awk -v n="$TOTAL" -v t="$ELAPSED" \
    'BEGIN { if (t > 0) printf "%.2f", n/t; else print "0" }')

AVG_LATENCY=$(awk -F, '
    $1 ~ /^[0-9]+$/ {
        sum += $2
        n++
    }
    END {
        if (n > 0) printf "%.3f", sum/n*1000
        else print "0"
    }
' "$TEMP_FILE")

TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

echo
echo "Benchmark Results"
echo "-----------------"
echo "Algorithm:       $ALGORITHM"
echo "Responses:       $TOTAL"
echo "Accepted:        $ACCEPTED"
echo "Rejected:        $REJECTED"
echo "Other:           $OTHER"
echo "Elapsed:         $ELAPSED seconds"
echo "Throughput:      $THROUGHPUT responses/second"
echo "Average latency: $AVG_LATENCY ms"

echo "$TIMESTAMP,$ALGORITHM,$TOTAL,$ACCEPTED,$REJECTED,$OTHER,$ELAPSED,$THROUGHPUT,$AVG_LATENCY" \
    >> "$RESULT_FILE"

echo
echo "Results appended to $RESULT_FILE"