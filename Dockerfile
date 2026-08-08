# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# Build argument — обов'язково передається через docker build --build-arg ENV=local
ARG ENV

WORKDIR /app

# Копіюємо go.mod (і go.sum якщо є)
COPY go.mod ./

# Завантажуємо залежності
RUN go mod download

# Копіюємо весь вихідний код
COPY . .

# Перевіряємо, що ENV передано і не порожнє — інакше build падає
RUN if [ -z "${ENV}" ]; then \
        echo "❌ Build error: ENV is not set. Use --build-arg ENV=<value>" >&2; \
        exit 1; \
    fi

# Збираємо статичний бінарник з підстановкою змінної env
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.env=${ENV} -s -w" \
    -o app .

# ─── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# Копіюємо лише бінарник зі stage builder
COPY --from=builder /app/app .

# Порт, який слухає сервер
EXPOSE 8080

# Запуск
CMD ["./app"]
