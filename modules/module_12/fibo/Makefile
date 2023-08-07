CMD = ./cmd/server
CONFIG = ./config/config.yaml

.PHONY: serve
serve:
	go run $(CMD)/main.go --config=$(CONFIG)

.PHONY: build
build:
	docker build -t yuki:5000/fibo-svc:v0.0.1 .
	docker push yuki:5000/fibo-svc:v0.0.1

