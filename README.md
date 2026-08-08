# golang_hello_world

Go-застосунок, що перевіряє доступність одного або кількох сайтів (HTTP-статус + час відповіді) і виводить результат у stdout (та опційно в лог-файл). Підтримує два режими виконання — керуються лише змінними оточення, окремих скриптів більше немає.

## Режими виконання

### 1. CronJob model (за замовчуванням, без `INTERVAL`)

Один прохід перевірки — і вихід. Саме цього контракту очікує Kubernetes CronJob: Pod повинен сам завершитися (`Completed`), інакше CronJob-контролер ніколи не побачить реального результату тика, а Job просто висітиме "Running" до таймауту.

Завершується з `exit 0`, якщо всі перевірені сайти відповіли 2xx/3xx, або `exit 1`, якщо хоч один — ні (unreachable, timeout, 4xx/5xx). Цей код виходу — те, що платформа читає для реального статусу "Succeeded"/"Failed" останнього запуску.

### 2. Continuous loop model (з `INTERVAL`)

Той самий tick повторюється кожні `INTERVAL` секунд нескінченно (не виходить сам) — для запуску як звичайний довготривалий Deployment, а не CronJob.

## Вимоги

- Go 1.22+
- Docker (опційно, для запуску в контейнері)

## Змінні оточення

| Змінна | Обов'язкова | За замовчуванням | Опис |
| --- | --- | --- | --- |
| `ENV` | так | — | Назва оточення. Значення довільне (`local`, `sandbox`, `prod` тощо) — обов'язкова лише непустота. Без неї додаток **не запуститься** (exit 1). |
| `URL_HOST` | ні | — (перевіряються 5 сайтів за замовчуванням) | Один конкретний сайт для перевірки замість стандартної п'ятірки. Приймає і домен (`google.com`), і повний URL (`https://google.com`) — якщо схему не вказано, підставляється `https://`. |
| `INTERVAL` | ні | — (one-shot, CronJob-режим) | Позитивне ціле число секунд — вмикає continuous loop model замість one-shot. |
| `LOG_FILE` | ні | — (лише stdout) | Шлях до файлу — кожен рядок виводу додатково дописується (append) і туди, і в stdout. |

Без `URL_HOST` перевіряються 5 вбудованих сайтів: `google.com`, `github.com`, `cloudflare.com`, `microsoft.com`, `amazon.com`.

## Запуск локально (без Docker)

```bash
# Збірка
make build
# або напряму:
go build -o app .

# CronJob-режим: перевірить 5 сайтів за замовчуванням і вийде
ENV=local ./app

# Перевірка одного конкретного сайту
ENV=local URL_HOST=google.com ./app

# Continuous loop: перевірка кожні 30с, лог і в stdout, і у файл
ENV=local INTERVAL=30 LOG_FILE=./checks.log ./app
```

Приклад виводу (CronJob-режим):

```text
env: local Hello this is local env
[2026-08-08 16:15:38] 🕒 Перевірка 5 сайт(ів)
[2026-08-08 16:15:38] ✅ https://google.com -> HTTP 200 (269ms)
[2026-08-08 16:15:38] ✅ https://github.com -> HTTP 200 (149ms)
[2026-08-08 16:15:39] ✅ https://cloudflare.com -> HTTP 200 (1.024s)
[2026-08-08 16:15:39] ✅ https://www.microsoft.com -> HTTP 200 (419ms)
[2026-08-08 16:15:39] ✅ https://amazon.com -> HTTP 200 (697ms)
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

# CronJob-режим - контейнер сам завершиться
docker run --rm -e ENV=local golang-hello-world

# Continuous loop у контейнері
docker run --rm -e ENV=local -e INTERVAL=30 golang-hello-world
```

Без `-e ENV=...` контейнер одразу впаде з помилкою (exit 1). Немає `-p`/`EXPOSE` — цей контейнер нічого не слухає.

## Makefile

```bash
make build            # зібрати бінарник app
make run ENV=local    # зібрати та запустити (CronJob-режим: одноразова перевірка + exit)
make clean            # видалити зібраний бінарник
```
