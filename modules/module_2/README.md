## 模块二作业


### 题目
```
编写一个 HTTP 服务器，大家视个人不同情况决定完成到哪个环节，但尽量把 1 都做完：

接收客户端 request，并将 request 中带的 header 写入 response header
读取当前系统的环境变量中的 VERSION 配置，并写入 response header
Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
当访问 localhost/healthz 时，应返回 200
```

### 答案
* 程序：modules/module_2/server
* 测试：modules/module_2/server_test.go

### 运行
```bash
$ make test
go test -run TestServer
HTTP SERVER: 
        Path      : /healthz
        ClientIP  : 127.0.0.1:54796
        StatusCode: 200

2023/06/05 08:17:32 
StatusCode is 200
Response Header is {
        "Accept-Encoding": [
                "gzip"
        ],
        "Content-Length": [
                "0"
        ],
        "Date": [
                "Mon, 05 Jun 2023 08:17:32 GMT"
        ],
        "My_header": [
                "hello"
        ],
        "User-Agent": [
                "Go-http-client/1.1"
        ],
        "Version": [
                "v1"
        ]
}
PASS
ok      module2 1.013s
```
可以看到
1. request中自定义的header "My_header"以及"Version"均被成功写入response header
2. 通过response header中的Version值与通过os.Setenv写入的环境变量的值一致
3. http server端有打印出客户端IP， HTTP响应码
4. 从响应体中获取的响应码为200