# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копіюємо go.mod (і go.sum якщо є)
COPY go.mod ./

# Завантажуємо залежності
RUN go mod download

# Копіюємо весь вихідний код
COPY . .

# ENV більше не потрібна на етапі build — читається в рантаймі з os.Getenv("ENV").
# Передавай її при запуску: docker run -e ENV=local ...
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o app .

# ─── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# Копіюємо лише бінарник зі stage builder
COPY --from=builder /app/app .

# CronJob execution model - цей бінарник відпрацьовує перевірку(и) і сам
# завершується (exit 0/1) щотика, тому тут немає жодного порту, який
# треба слухати/EXPOSE - на відміну від довготривалого веб-сервера.
CMD ["./app"]
