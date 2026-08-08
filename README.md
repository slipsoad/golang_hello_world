# golang_hello_world

Простий Go веб-сервер, що показує локальну/публічну IP-адресу та значення оточення (`ENV`).

## Вимоги

- Go 1.22+
- Docker (опційно, для запуску в контейнері)

## Змінні оточення

| Змінна | Обов'язкова | За замовчуванням | Опис |
|---|---|---|---|
| `ENV` | так | — | Назва оточення (`local`, `dev`, `prod` тощо). Без неї додаток **не запуститься** (падає з exit 1). |
| `PORT` | ні | `8080` | Порт, на якому слухає HTTP-сервер. |

## Запуск локально (без Docker)

```bash
# Збірка
make build
# або напряму:
go build -o app .

# Запуск (ENV обов'язковий)
make run ENV=local
# або напряму:
ENV=local ./app
```

Без `ENV` запуск завершиться помилкою:

```
❌ Startup error: змінна оточення 'ENV' не встановлена.
   Використовуйте: ENV=local ./app
   Або: docker run -e ENV=local ...
```

## Запуск через Docker

```bash
# Збірка образу (ENV НЕ потрібен на цьому кроці)
docker build -t golang-hello-world .

# Запуск (ENV обов'язковий саме тут, при docker run)
docker run --rm -p 8080:8080 -e ENV=local golang-hello-world
```

Без `-e ENV=...` контейнер одразу впаде з помилкою.

## Endpoints

| Метод | Шлях | Опис |
|---|---|---|
| GET | `/` | Головна HTML-сторінка |
| GET | `/api/ip` | JSON з локальною/публічною IP та `env` |
| GET | `/health` | Health check (`status`, `env`, `timestamp`) |

Приклад:

```bash
curl http://localhost:8080/health
```

## Допоміжні скрипти (scripts/)

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
make run ENV=local    # зібрати та запустити з ENV=local
make clean            # видалити зібраний бінарник
```
