# 模块三作业


## 题目
*  构建本地镜像
*  编写 Dockerfile 将模块二作业编写的 httpserver 容器化
*  将镜像推送至 docker 官方镜像仓库
*  通过 docker 命令本地启动 httpserver
*  通过 nsenter 进入容器查看 IP 配置
*  作业需编写并提交 Dockerfile 及源代码。

## 答案
### 1. 构建本地镜像 && 编写 Dockerfile 将模块二作业编写的 httpserver 容器化

通过二阶段构建将httpserver打包为镜像, 使用Dockerfile(`modules/module_3/Dockerfile`)如下：
```Dockerfile
# stage0: 编译go可执行文件
FROM golang:alpine as builder

WORKDIR /home

COPY ./ .

ENV CMD ./cmd/server

RUN CGO_ENABLED=0 go build -o ${CMD}/server ${CMD}/

# stage1: 将stage0中编译完毕的可执行文件放置于alpine容器中并启动
FROM alpine:3.18

WORKDIR /home

COPY --from=builder /home/cmd/server/server ./server

CMD ./server
```

在根目录执行`make build`（即执行`docker build -t yukichen7221/http-server-demo:v0.0.1 .`）之后，可以通过docker images查看到镜像`yukichen7221/http-server-demo:v0.0.1`被成功构建
```bash
$ docker images
REPOSITORY                         TAG                                        IMAGE ID       CREATED          SIZE
yukichen7221/http-server-demo      v0.0.1                                     0cccc772a13f   10 minutes ago   14MB
```
说明容器构建成功

### 2.将镜像推送至 docker 官方镜像仓库

* 登录`hub.docker.com`，并创建repository [yukichen7221
/
http-server-demo](https://hub.docker.com/repository/docker/yukichen7221/http-server-demo/general)

* 执行`docker login`本地登录`hub.docker.com`
* 推送镜像
    ```bash
    $ docker push yukichen7221/http-server-demo:v0.0.1
    ```
    随后可以在新建的[repository](https://hub.docker.com/repository/docker/yukichen7221/http-server-demo/general)中查看到刚刚推送的`v0.0.1`版本的镜像

### 3.通过 docker 命令本地启动 httpserver

启动容器
```bash
$ docker run -d -p 8080:8080 --rm --name serverdemo yukichen7221/http-server-demo:v0.0.1
```

在`modules/module_3/server_test.go`中编写测试单元测试是否能够请求到部署的容器服务

```bash
$ go test --count 1 . -v
=== RUN   TestServer
--- PASS: TestServer (0.00s)
PASS
ok      module3 0.021s
```

请求通过，说明容器中的服务已经正常启动

进一步查看容器中打印的日志：
```bash
$ docker logs serverdemo
HTTP SERVER: 
        Path      : /healthz
        ClientIP  : 192.168.245.1:52600
        StatusCode: 200
```
可以看到日志也随着单元测试的请求产生了，工作正常

### 4.通过 nsenter 进入容器查看 IP 配置
获取容器pid
```bash
$ docker inspect --format '{{.State.Pid}}' serverdemo
```
使用nsenter查看IP配置
```bash
$ nsenter -t $(docker inspect --format '{{.State.Pid}}' serverdemo) -n ifconfig
eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 172.17.0.5  netmask 255.255.0.0  broadcast 172.17.255.255
        ether 02:42:ac:11:00:05  txqueuelen 0  (Ethernet)
        RX packets 78  bytes 5417 (5.4 KB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 56  bytes 4020 (4.0 KB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

lo: flags=73<UP,LOOPBACK,RUNNING>  mtu 65536
        inet 127.0.0.1  netmask 255.0.0.0
        loop  txqueuelen 1000  (Local Loopback)
        RX packets 0  bytes 0 (0.0 B)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 0  bytes 0 (0.0 B)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
```

验证是否与直接执行容器命令的结果一致：
```bash
$ docker exec serverdemo ifconfig
eth0      Link encap:Ethernet  HWaddr 02:42:AC:11:00:05  
          inet addr:172.17.0.5  Bcast:172.17.255.255  Mask:255.255.0.0
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:78 errors:0 dropped:0 overruns:0 frame:0
          TX packets:56 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:0 
          RX bytes:5417 (5.2 KiB)  TX bytes:4020 (3.9 KiB)

lo        Link encap:Local Loopback  
          inet addr:127.0.0.1  Mask:255.0.0.0
          UP LOOPBACK RUNNING  MTU:65536  Metric:1
          RX packets:0 errors:0 dropped:0 overruns:0 frame:0
          TX packets:0 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000 
          RX bytes:0 (0.0 B)  TX bytes:0 (0.0 B)
```
可以看到，eth0中的ip配置基本是一致的，说明刚刚使用nsenter检查的结果是正确的