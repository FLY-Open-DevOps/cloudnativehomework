# 模块八作业


## 题目
1. 第一部分

    现在你对 Kubernetes 的控制面板的工作机制是否有了深入的了解呢？
    是否对如何构建一个优雅的云上应用有了深刻的认识，那么接下来用最近学过的知识把你之前编写的 http 以优雅的方式部署起来吧，你可能需要审视之前代码是否能满足优雅上云的需求。

    作业要求：编写 Kubernetes 部署脚本将 httpserver 部署到 Kubernetes 集群，以下是你可以思考的维度。

    * 优雅启动
    * 优雅终止
    * 资源需求和 QoS 保证
    * 探活
    * 日常运维需求，日志等级
    * 配置和代码分离

2. 第二部分

    除了将 httpServer 应用优雅的运行在 Kubernetes 之上，我们还应该考虑如何将服务发布给对内和对外的调用方。

    来尝试用 Service, Ingress 将你的服务发布给集群外部的调用方吧。
    在第一部分的基础上提供更加完备的部署 spec，包括（不限于）：

    * Service
    * Ingress
    
    可以考虑的细节

    * 如何确保整个应用的高可用。
    * 如何通过证书保证 httpServer 的通讯安全。


## 答案


> 所有关于kuberentes的配置文件得路径均在`/manifest`中

### http server改造
为了满足题目需求，对原本的http server做了以下改造

1. 新增了一个计算Fibonacci数列的服务，路径为`/fibo`
    * 计算数列中第n个数的值，对应实现在`/internal/module8/fibo.go`
    * 使用哈希表作为缓存，如果缓存中存有结果，则直接返回，否则进行计算后，在返回接过前，将结果存放到哈希表中，对应实现在`/internal/module8/cache.go`
    * 哈希表在DEV模式下打印DEBUG level的日志，如：
        ```bash
        time=2023-07-21T06:41:29.609Z level=DEBUG msg="cache miss" key=30
        time=2023-07-21T06:41:29.620Z level=DEBUG msg="cache fibo" key=30 value=832040
        ```
        在PROD模式下则不打印DEBUG level的日志，仅输出INFO level以上的日志，ru：
        ```bash
        time=2023-07-21T06:50:13.816Z level=INFO msg="server start" addr=0.0.0.0:8080
        time=2023-07-21T06:50:19.277Z level=INFO msg=start id=6d891599-6b46-4633-9267-a13bc8169de6 path=/dev/fibo client=127.0.0.1:33494
        time=2023-07-21T06:50:19.278Z level=ERROR msg="fibo caculate failed" errmsg="strconv.Atoi: parsing \"30x\": invalid syntax"
        time=2023-07-21T06:50:19.278Z level=INFO msg=end id=6d891599-6b46-4633-9267-a13bc8169de6 path=/dev/fibo client=127.0.0.1:33494 "status code"=400 duration=733.585µs
        ```
    * 日志使用[slog](golang.org/x/exp/slog)实现
    * 可以计算的最大值根据由配置文件中的maxseq字段决定，配置文件可参考`/config/config.yaml`，如果传入参数大于该数字，则返回500（InternalServerError）

2. 通过启动时配置的环境来决定请求的url
    * 比如配置环境中的env字段为DEV，则`/healthz`路径的实际请求路径应为`/dev/healthz`，env字段为PROD时，则`/healthz`路径的实际请求路径应为`/prod/healthz`

3. 为了配合优雅终止，server持续监听SIGTERM以及SIGINT两个信号，如果接收到两个信号的时候，则先将http server终止，然后再退出进程，实现如下（具体实现在`/cmd/main.go`）：
    ```go
    func main() {
        // omit login of reading configuration and starting server

        ch := make(chan os.Signal, 1)
        signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
        <-ch
        if err := server.Stop(ctx); err != nil {
            log.Fatal(err)
            os.Exit(1)
        }
        os.Exit(0)
    }

    ```
    例如在接收到SIGTERM信号时，程序行为是：
    ```bash
    time=2023-07-21T07:05:09.216Z level=INFO msg="graceful stopping server"
    2023/07/21 07:05:09 server down because: http: Server closed
    ```
    打印关闭信息，然后再退出进程

### 优雅启动 && 探活
在Deployment文件中配置readiness Probe以及liveness Probe，使用`/healthz`路径配置为这两个探针的请求路径，具体如下：
```yaml
livenessProbe:
httpGet:
    path: /dev/healthz
    port: server-port
initialDelaySeconds: 1
periodSeconds: 5
readinessProbe:
httpGet:
    path: /dev/healthz
    port: server-port
initialDelaySeconds: 1
periodSeconds: 2
```
上面是DEV环境下的两个探针的配置，PROD环境则将对应路径中的替换为prod即可，后续应使用HELM模板进行统一配置

### 优雅终止
server持续监听SIGTERM以及SIGINT两个信号，如果接收到两个信号的时候，则先将http server终止，然后再退出进程。Kubernetes在终止Pod中的容器时，会给容器中的进程先发送SIGTERM信号，容器进程监听到该信号便会主动关闭http server并退出


### 资源需求和 QoS 保证
在Deployment文件中为容器配置Burstable的资源声明，如下：
```yaml
resources:
    requests:
        memory: "128Mi"
        cpu: "100m"
    limits:
        memory: "1Gi"
        cpu: "1"
```

### 日常运维需求，日志等级
日志均输出到控制台中，根据配置文件中定义env字段决定容器打印的日志等级，如果是DEV环境，则打印DEBUG level以上的日志，在PROD环境，则打印INFO level以上的日志。后续改造时需要配置FileBeat等组件通过Sidecar模式或者DaemonSet模式采集Pod中的控制台输出，并采集到ElasticSearch中，通过Kibana进行查找与分析


### 配置和代码分离
将配置写入到ConfigMap中，如下
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: http-server-demo-dev
  namespace: cloudnative
data: 
  config.yaml: |
    env: DEV
    maxseq: 40
    cacheresult: true
    port: 80
```

在容器启动时，将config.yaml字段挂载到容器中，由应用程序进行读取以获得配置信息，如下：
```yaml
spec:
    containers:
        volumeMounts:
        - name: config
        mountPath: "/home/config"
        readOnly: true
    volumes:
    - name: config
        configMap:
        name: http-server-demo-dev
```
应用程序根据启动命令的传入参数获取到配置文件的路径，来读取配置文件，具体实现在`/cmd/main.go`中：
```go
func main() {
	ctx := context.Background()
	cfgdir := flag.String("config", "./config.yaml", "dir of configuration")
	flag.Parse()
	cfg := cfgmaker.New(&module8.Config{}).
		ReadFromYamlFile(*cfgdir).
		Get().(*module8.Config)
    // omit rest of logic
}

```

### 如何确保整个应用的高可用
为应用配置多个副本，具体在Deployment文件中进行如下设置：
```yaml
spec:
  replicas: 3
```
为http server这个无状态应用配置3个副本，如果其中一个副本下线了，在该副本重新上线之前还有剩下的两个副本可以正常提供服务，不至于服务完全瘫痪


### 配置Service以及Ingress，以及通过证书保证 httpServer 的通讯安全
首先是配置Service：
```yaml
apiVersion: v1
kind: Service
metadata:
  name: http-server-demo-dev
  namespace: cloudnative
spec:
  selector:
    app: http-server-demo-dev
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
```

然后是生成证书，使用Makefile中的certgen指令进行生成，生成的证书会在路径`/ca`下
```bash
$  make certgen
openssl genrsa -out /root/playground/geekbang/cloudnative/modules/module_8/ca/cloudnative.key 4096
Generating RSA private key, 4096 bit long modulus (2 primes)
.......................................................................................................................++++
...........................................................................................................................................................................................................................++++
e is 65537 (0x010001)
openssl req -new -x509 -days 3650 \
        -subj "/C=GB/L=China/O=grpc-server/CN=cloudnative" \
        -addext "subjectAltName = DNS:cloudnative" \
        -key /root/playground/geekbang/cloudnative/modules/module_8/ca/cloudnative.key -out /root/playground/geekbang/cloudnative/modules/module_8/ca/cloudnative.crt
```

将证书配置到Kubernetes集群中的Secret中
```bash
$ kubectl create secret -n cloudnative tls cloudnative --cert=./ca/cloudnative.crt --key=./ca/cloudnative.key
```

然后配置Ingress，使用路径来区分DEV和PROD两个环境的服务，并在tls字段中为服务配置相应的证书
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: http-server-demo
  namespace: cloudnative
spec:
  tls:
    - hosts:
        - cloudnative
      secretName: cloudnative
  rules:
    - http:
        paths:
          - path: /dev
            pathType: Prefix
            backend:
              service:
                name: http-server-demo-dev
                port:
                  number: 80
          - path: /prod
            pathType: Prefix
            backend:
              service:
                name: http-server-demo-prod
                port:
                  number: 80
```

### 总体流程

1. 编译程序与镜像构建发布
    这里搭建了本地的registry仓库，将镜像推送到该仓库中
    ```bash
    $ make build
    docker build -t yuki:5000/http-server-demo:v0.0.3 .
    Sending build context to Docker daemon  35.33kB
    Step 1/9 : FROM golang:alpine as builder
    ---> e9f54d6722ab
    Step 2/9 : WORKDIR /home
    ---> Using cache
    ---> 3489dfb7b972
    Step 3/9 : COPY ./ .
    ---> 4a9a126fb0a3
    Step 4/9 : ENV CMD ./cmd/server
    ---> Running in 7fc85ead65fb
    Removing intermediate container 7fc85ead65fb
    ---> 8bedafe9d5cb
    Step 5/9 : RUN CGO_ENABLED=0 GOPROXY=https://mirrors.aliyun.com/goproxy,https://goproxy.io,direct go build -o ${CMD}/server ${CMD}/
    ---> Running in 2fd7348de6d3
    go: downloading github.com/yukiouma/cfg-maker v0.1.0
    go: downloading github.com/google/uuid v1.3.0
    go: downloading golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1
    go: downloading gopkg.in/yaml.v3 v3.0.1
    Removing intermediate container 2fd7348de6d3
    ---> 5929d1348b3c
    Step 6/9 : FROM alpine:3.18
    ---> 5e2b554c1c45
    Step 7/9 : WORKDIR /home
    ---> Using cache
    ---> 1c2d67ce9742
    Step 8/9 : COPY --from=builder /home/cmd/server/server ./server
    ---> a7a3e6c67568
    Step 9/9 : CMD ./server
    ---> Running in b6a071cf0cbd
    Removing intermediate container b6a071cf0cbd
    ---> de8325e1457f
    Successfully built de8325e1457f
    Successfully tagged yuki:5000/http-server-demo:v0.0.3
    docker push yuki:5000/http-server-demo:v0.0.3
    The push refers to repository [yuki:5000/http-server-demo]
    3babec273e2e: Pushed 
    bb01bd7e32b5: Layer already exists 
    v0.0.3: digest: sha256:2bc2427596f64dd5fbd1b67a9c593584578ba28ace76abe54ddf5a843f1d8107 size: 739
    ```
2. 发布到Kuberntes中
    首先生成证书与将证书配置到Kubernetes集群中的Secret中
    ```bash
    $ make certgen
    openssl genrsa -out /root/playground/geekbang/cloudnative/modules/module_8/ca/cloudnative.key 4096
    Generating RSA private key, 4096 bit long modulus (2 primes)
    .......................................................................................................................++++
    ...........................................................................................................................................................................................................................++++
    e is 65537 (0x010001)
    openssl req -new -x509 -days 3650 \
            -subj "/C=GB/L=China/O=grpc-server/CN=cloudnative" \
            -addext "subjectAltName = DNS:cloudnative" \
            -key /root/playground/geekbang/cloudnative/modules/module_8/ca/cloudnative.key -out /root/playground/geekbang/cloudnative/modules/module_8/ca/cloudnative.crt

    $ kubectl create secret -n cloudnative tls cloudnative --cert=./ca/cloudnative.crt --key=./ca/cloudnative.key
    ```
3. 发布到Kubernetes集群
    ```bash
    $ kubectl apply -f ./manifest/
    configmap/http-server-demo-dev created
    configmap/http-server-demo-prod created
    deployment.apps/http-server-demo-dev created
    deployment.apps/http-server-demo-prod created
    ingress.networking.k8s.io/http-server-demo created
    service/http-server-demo-dev created
    service/http-server-demo-prod created
    ```

    查看容器状态
    ```bash
    $ kubectl get pods -n cloudnative
    NAME                                     READY   STATUS    RESTARTS   AGE
    http-server-demo-dev-6c7678cdcb-7vnhp    1/1     Running   0          51s
    http-server-demo-dev-6c7678cdcb-8m2bx    1/1     Running   0          51s
    http-server-demo-dev-6c7678cdcb-jpjc2    1/1     Running   0          51s
    http-server-demo-prod-7557974d56-9l7hx   1/1     Running   0          51s
    http-server-demo-prod-7557974d56-j4vgv   1/1     Running   0          51s
    http-server-demo-prod-7557974d56-x56mw   1/1     Running   0          51s   
    ```

    可以看到容器均在正常工作

4. 验证
    查看集群地址：
    ```bash
    $ minikube ip
    192.168.49.2
    $ nslookup cloudnative
    Server:         127.0.0.53
    Address:        127.0.0.53#53

    Non-authoritative answer:
    Name:   cloudnative
    Address: 192.168.49.2
    ```
    尝试在集群外部对两个环境的服务发起请求
    ```bash
    $ curl https://cloudnative/dev/fibo?n=10
    {"result": 55} 
    $ curl https://cloudnative/prod/fibo?n=10
    {"result": 55}
    ```
    均能成功

    查看两个环境的日志信息：
    ```bash
    $ kubectl logs http-server-demo-dev-6c7678cdcb-8m2bx -n cloudnative
    time=2023-07-21T07:41:25.334Z level=INFO msg=start id=4f9ec995-3856-42e4-990d-52fff5c85947 path=/dev/fibo client=172.17.0.2:53954
    time=2023-07-21T07:41:25.334Z level=DEBUG msg="cache miss" key=10
    time=2023-07-21T07:41:25.334Z level=DEBUG msg="cache fibo" key=10 value=55
    time=2023-07-21T07:41:25.334Z level=INFO msg=end id=4f9ec995-3856-42e4-990d-52fff5c85947 path=/dev/fibo client=172.17.0.2:53954 "status code"=200 duration=128.804µs

    $ kubectl logs http-server-demo-prod-7557974d56-j4vgv -n cloudnative
    time=2023-07-21T07:41:32.325Z level=INFO msg=start id=26990a9d-f576-4281-b7fb-be01e5eec278 path=/prod/fibo client=172.17.0.2:53068
    time=2023-07-21T07:41:32.325Z level=INFO msg=end id=26990a9d-f576-4281-b7fb-be01e5eec278 path=/prod/fibo client=172.17.0.2:53068 "status code"=200 duration=160.097µs
    ```
    可以看到，在两个完全相同输入的请求中，在DEV环境中存在level=DEBUG的日志信息，而在PROD环境中仅又level=INFO的日志信息

自此，服务发布到Kubernetes中并完成了对外访问的暴露