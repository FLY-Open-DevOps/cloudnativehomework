CMD = ./cmd/server
CONFIG = ./config/config.yaml

.PHONY: serve
serve:
	go run $(CMD)/main.go --config=$(CONFIG)

.PHONY: build
build:
	docker build -t yuki:5000/http-server-demo:v0.0.3 .
	docker push yuki:5000/http-server-demo:v0.0.3

.PHONY: certgen
certgen:
	openssl genrsa -out $(PWD)/ca/cloudnative.key 4096
	openssl req -new -x509 -days 3650 \
    	-subj "/C=GB/L=China/O=grpc-server/CN=cloudnative" \
    	-addext "subjectAltName = DNS:cloudnative" \
    	-key $(PWD)/ca/cloudnative.key -out $(PWD)/ca/cloudnative.crt