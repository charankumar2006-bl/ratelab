#!/bin/bash

DURATION=30
INTERVAL=1
OUTPUT_FILE="resource_results.csv"

echo "timestamp,cpu_percent,memory_percent,memory_used_kb" > "$OUTPUT_FILE"

echo "=================================="
echo "RateLab Resource Monitoring"
echo "=================================="
echo "Duration: $DURATION seconds"
echo "Sampling interval: $INTERVAL second"
echo

# Find the gateway process.
GATEWAY_PID=$(pgrep -n -x gateway)

if [ -z "$GATEWAY_PID" ]; then
    echo "Gateway process not found."
    echo "Start the gateway before running this script."
    exit 1
fi

echo "Monitoring gateway PID: $GATEWAY_PID"
echo "timestamp,cpu_percent,memory_percent,memory_used_kb" > "$OUTPUT_FILE"

END_TIME=$((SECONDS + DURATION))

while [ "$SECONDS" -lt "$END_TIME" ]; do
    if ! kill -0 "$GATEWAY_PID" 2>/dev/null; then
        echo "Gateway process stopped."
        break
    fi

    TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

    STATS=$(ps -p "$GATEWAY_PID" -o %cpu=,%mem=,rss=)

    if [ -n "$STATS" ]; then
        read -r CPU MEMORY RSS <<< "$STATS"
        echo "$TIMESTAMP,$CPU,$MEMORY,$RSS" >> "$OUTPUT_FILE"
        echo "CPU: $CPU% | Memory: $MEMORY% | RSS: ${RSS}KB"
    fi

    sleep "$INTERVAL"
done

echo
echo "Resource monitoring complete."
echo "Results saved to $OUTPUT_FILE"