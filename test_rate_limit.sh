#!/bin/bash

BASE_URL="http://localhost:8080"
CLIENT_ID="alice"

echo "RateLab Fixed Window Experiment"
echo "Client: $CLIENT_ID"
echo "----------------------------------------"

for i in {1..6}; do
    echo
    echo "Request $i"

    curl -s -i \
        -H "X-Client-ID: $CLIENT_ID" \
        "$BASE_URL/api/data"
done

echo
echo "----------------------------------------"
echo "Experiment completed."
