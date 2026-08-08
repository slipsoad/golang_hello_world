# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# Build argument — передається через docker build --build-arg ENV=local
ARG ENV=local

WORKDIR /app

# Копіюємо go.mod (і go.sum якщо є)
COPY go.mod ./

# Завантажуємо залежності
RUN go mod download

# Копіюємо весь вихідний код
COPY . .

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
