.PHONY: bff
bff:
	cd ./bff && make serve

.PHONY: caculator
caculator:
	cd ./caculator && make serve

.PHONY: fibo
fibo:
	cd ./fibo && make serve

.PHONY: build
build:
	cd /root/playground/geekbang/cloudnative/modules/module_12/fibo && make build
	cd /root/playground/geekbang/cloudnative/modules/module_12/caculator && make build
	cd /root/playground/geekbang/cloudnative/modules/module_12/bff && make build


.PHONY: certgen
certgen:
	openssl genrsa -out $(PWD)/ca/cloudnative.io.key 4096
	openssl req -new -x509 -days 3650 \
    	-subj "/C=GB/L=China/O=grpc-server/CN=cloudnative.io" \
		-addext "subjectAltName = DNS:cloudnative" \
    	-key $(PWD)/ca/cloudnative.io.key -out $(PWD)/ca/cloudnative.io.crt