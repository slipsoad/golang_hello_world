#!/bin/sh
# Простий scheduler для перевірки роботи додатка.
# Періодично (кожні INTERVAL секунд) звертається до /health і логує статус.
#
# Usage:
#   ./scripts/scheduler_check.sh
#   URL=http://localhost:8080/health INTERVAL=5 ./scripts/scheduler_check.sh

URL="${URL:-http://localhost:8080/health}"
INTERVAL="${INTERVAL:-10}"

echo "🕒 Scheduler запущено. URL=$URL, INTERVAL=${INTERVAL}s"

while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    response=$(curl -s -o /tmp/scheduler_check_body -w "%{http_code}" "$URL")

    if [ "$response" = "200" ]; then
        body=$(cat /tmp/scheduler_check_body)
        echo "[$timestamp] ✅ OK (HTTP $response) - $body"
    else
        echo "[$timestamp] ❌ FAIL (HTTP $response)"
    fi

    sleep "$INTERVAL"
done
