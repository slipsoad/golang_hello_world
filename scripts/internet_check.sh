#!/bin/sh
# Простий scheduler для перевірки наявності інтернету (та, опційно, конкретного сайту).
# Періодично (кожні INTERVAL секунд) перевіряє з'єднання і пише результат
# одночасно в stdout та в лог-файл.
#
# Usage:
#   ./scripts/internet_check.sh
#   HOST=8.8.8.8 INTERVAL=5 LOG_FILE=./internet.log ./scripts/internet_check.sh
#
# Перевірка конкретного сайту (замість ping на HOST) — задай URL_HOST:
#   URL_HOST=google.com ./scripts/internet_check.sh
#   URL_HOST=https://google.com ./scripts/internet_check.sh
# Формат довільний — можна писати як просто домен, так і повний URL зі схемою;
# якщо схему не вказано, скрипт сам підставить https://.

HOST="${HOST:-8.8.8.8}"
URL_HOST="${URL_HOST:-}"
INTERVAL="${INTERVAL:-10}"
LOG_FILE="${LOG_FILE:-$(dirname "$0")/internet_check.log}"

if [ -n "$URL_HOST" ]; then
    case "$URL_HOST" in
        http://*|https://*) TARGET_URL="$URL_HOST" ;;
        *) TARGET_URL="https://$URL_HOST" ;;
    esac
    echo "🕒 Internet check запущено. URL_HOST=$TARGET_URL, INTERVAL=${INTERVAL}s, LOG_FILE=$LOG_FILE"
else
    echo "🕒 Internet check запущено. HOST=$HOST, INTERVAL=${INTERVAL}s, LOG_FILE=$LOG_FILE"
fi

while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    if [ -n "$URL_HOST" ]; then
        http_code=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$TARGET_URL")
        if [ "$http_code" -ge 200 ] 2>/dev/null && [ "$http_code" -lt 400 ] 2>/dev/null; then
            line="[$timestamp] ✅ Сайт $TARGET_URL працює (HTTP $http_code)"
        else
            line="[$timestamp] ❌ Сайт $TARGET_URL не відповідає (HTTP $http_code)"
        fi
    else
        if ping -c 1 -W 2 "$HOST" >/dev/null 2>&1; then
            line="[$timestamp] ✅ Інтернет є (ping $HOST успішний)"
        else
            line="[$timestamp] ❌ Інтернету немає (ping $HOST не пройшов)"
        fi
    fi

    echo "$line" | tee -a "$LOG_FILE"

    sleep "$INTERVAL"
done
