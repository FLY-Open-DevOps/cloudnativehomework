CMD = ./cmd/server

.PHONY: serve
serve:
	go run $(CMD)/main.go

.PHONY: build
build:
	docker build -t yukichen7221/http-server-demo:v0.0.1 .

.PHONY: deploy
deploy:
	docker run -d -p 8080:8080 --rm --name serverdemo yukichen7221/http-server-demo:v0.0.1

.PHONY: test
test:
	go test --count 1 . -v