FROM golang:alpine as builder

WORKDIR /home

COPY ./ .

ENV CMD ./cmd/server

RUN CGO_ENABLED=0 GOPROXY=https://mirrors.aliyun.com/goproxy,https://goproxy.io,direct go build -o ${CMD}/server ${CMD}/


FROM alpine:3.18

WORKDIR /home

COPY --from=builder /home/cmd/server/server ./server

CMD ./server