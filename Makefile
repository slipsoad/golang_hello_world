## Usage:
##   make build             → builds the binary (ENV not required at build time)
##   make run ENV=local     → builds and runs with ENV=local
##   make run               → builds and runs; fails at startup (ENV not set)

.PHONY: build run clean

build:
	go build -o app .

run: build
	ENV=$(ENV) ./app

clean:
	rm -f app
