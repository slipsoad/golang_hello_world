# golang_hello_world

Go-застосунок для CronJob-моделі виконання: щотика перевіряє доступність одного або кількох сайтів (HTTP-статус + час відповіді), виводить результат у stdout і **завершується сам** (`exit 0` якщо всі перевірки пройшли, `exit 1` якщо хоч одна ні).

## CronJob execution model

На відміну від довготривалого веб-сервера, цей бінарник **не слухає жоден порт і не працює вічно** — він відпрацьовує перевірку(и) за один прохід і виходить. Це саме той контракт, якого очікує Kubernetes CronJob: Pod повинен сам завершитися (`Completed`), інакше CronJob-контролер ніколи не побачить реального результату тика (`Succeeded`/`Failed`), а Job просто висітиме "Running" до таймауту.

## Вимоги

- Go 1.22+
- Docker (опційно, для запуску в контейнері)

## Змінні оточення

| Змінна | Обов'язкова | За замовчуванням | Опис |
|---|---|---|---|
| `ENV` | так | — | Назва оточення. Значення довільне (`local`, `sandbox`, `prod` тощо) — обов'язкова лише непустота. Без неї додаток **не запуститься** (падає з exit 1). |
| `URL_HOST` | ні | — (перевіряються 5 сайтів за замовчуванням) | Один конкретний сайт для перевірки замість стандартної п'ятірки. Приймає і домен (`google.com`), і повний URL (`https://google.com`) — якщо схему не вказано, підставляється `https://`. |

Без `URL_HOST` перевіряються 5 вбудованих сайтів: `google.com`, `github.com`, `cloudflare.com`, `microsoft.com`, `amazon.com`.

## Запуск локально (без Docker)

```bash
# Збірка
make build
# або напряму:
go build -o app .

# Запуск (ENV обов'язковий) - перевірить 5 сайтів за замовчуванням і вийде
make run ENV=local
# або напряму:
ENV=local ./app

# Перевірка одного конкретного сайту
ENV=local URL_HOST=google.com ./app
```

Приклад виводу:

```text
env: local Hello this is local env
2026/08/08 16:15:38 🕒 CronJob tick - перевірка 5 сайт(ів)
✅ https://google.com -> HTTP 200 (269ms)
✅ https://github.com -> HTTP 200 (149ms)
✅ https://cloudflare.com -> HTTP 200 (1.024s)
✅ https://www.microsoft.com -> HTTP 200 (419ms)
✅ https://amazon.com -> HTTP 200 (697ms)
2026/08/08 16:15:41 ✅ CronJob tick завершено - усі перевірки пройшли
```

Без `ENV` запуск завершиться помилкою (exit 1):

```text
❌ Startup error: змінна оточення 'ENV' не встановлена.
   Використовуйте: ENV=local ./app
   Або: docker run -e ENV=local ...
```

## Запуск через Docker

```bash
# Збірка образу (ENV НЕ потрібен на цьому кроці)
docker build -t golang-hello-world .

# Запуск (ENV обов'язковий саме тут, при docker run) - контейнер сам завершиться
docker run --rm -e ENV=local golang-hello-world

# Перевірка одного конкретного сайту
docker run --rm -e ENV=local -e URL_HOST=github.com golang-hello-world
```

Без `-e ENV=...` контейнер одразу впаде з помилкою (exit 1). Немає `-p`/`EXPOSE` — цей контейнер нічого не слухає.

## Допоміжні скрипти (scripts/)

Це окремі bash-скрипти для локального/ручного моніторингу — незалежні від `main.go` та Docker-образу вище; вони працюють у нескінченному циклі (не для CronJob-режиму).

### `scheduler_check.sh` — періодична перевірка `/health`

```bash
./scripts/scheduler_check.sh
# з налаштуваннями:
URL=http://localhost:8080/health INTERVAL=5 ./scripts/scheduler_check.sh
```

Кожні `INTERVAL` секунд звертається до `/health` і виводить у stdout `✅ OK` або `❌ FAIL`.

### `internet_check.sh` — періодична перевірка наявності інтернету або конкретного сайту

```bash
./scripts/internet_check.sh
# з налаштуваннями:
HOST=1.1.1.1 INTERVAL=5 LOG_FILE=./internet.log ./scripts/internet_check.sh
```

Кожні `INTERVAL` секунд пінгує `HOST` (за замовчуванням `8.8.8.8`) і одночасно пише результат у stdout **і** у `LOG_FILE`.

Щоб перевіряти конкретний сайт (HTTP-запит замість ping), задай `URL_HOST`:

```bash
URL_HOST=google.com ./scripts/internet_check.sh
# або з повною схемою — обидва формати рівнозначні:
URL_HOST=https://google.com ./scripts/internet_check.sh
```

Формат довільний — можна вказати просто домен (`google.com`) або повний URL зі схемою (`https://google.com`); якщо схему не вказано, скрипт сам підставить `https://`. Коли `URL_HOST` заданий, він має пріоритет над `HOST`.

## Makefile

```bash
make build            # зібрати бінарник app
make run ENV=local    # зібрати та запустити з ENV=local (одноразова перевірка + exit)
make clean            # видалити зібраний бінарник
```
