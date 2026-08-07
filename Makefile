## Usage:
##   make build ENV=local   → builds with env=local
##   make build             → fails with an error (ENV not set)
##   make run ENV=local     → builds and runs

.PHONY: build run clean

build:
ifndef ENV
	$(error ❌ ENV is not set. Usage: make build ENV=local)
endif
	go build -ldflags "-X main.env=$(ENV)" -o app .

run: build
	./app

clean:
	rm -f app
