## 课后练习 1.1

### 题目
```
编写一个小程序：
给定一个字符串数组
[“I”,“am”,“stupid”,“and”,“weak”]
用 for 循环遍历该数组并修改为
[“I”,“am”,“smart”,“and”,“strong”]
```

### 答案
* 程序：modules/module_1/1-1.go
* 测试：modules/module_1/1-1_test.go

运行
```bash
$ cd modules/module_1 && make make test1-1
```
结果
```bash
$ make test1-1
go test -run Test1_1
2023/06/05 05:49:51 origin slice is [I am stupid and weak]
2023/06/05 05:49:51 modified slice is [I am smart and strong]
PASS
ok      module1 0.013s
```

## 课后练习 1.2

### 题目
```
基于 Channel 编写一个简单的单线程生产者消费者模型：

队列：
队列长度 10，队列元素类型为 int
生产者：
每 1 秒往队列中放入一个类型为 int 的元素，队列满时生产者可以阻塞
消费者：
每一秒从队列中获取一个元素并打印，队列为空时消费者阻塞
```

### 答案
* 程序：modules/module_1/1-2.go
* 测试：modules/module_1/1-2_test.go

运行
```bash
$ cd modules/module_1 && make make test1-2
```
结果
```bash
$ make test1-2
go test -run Test1_2
2023/06/05 06:17:03 Producer sent element: 2
2023/06/05 06:17:03 Consumer got element: 2
2023/06/05 06:17:04 Producer sent element: 78
2023/06/05 06:17:04 Consumer got element: 78
2023/06/05 06:17:05 Producer sent element: 72
2023/06/05 06:17:05 Consumer got element: 72
2023/06/05 06:17:06 Producer sent element: 53
2023/06/05 06:17:06 Consumer got element: 53
2023/06/05 06:17:07 Producer sent element: 86
2023/06/05 06:17:07 Consumer got element: 86
2023/06/05 06:17:08 Producer Terminated
2023/06/05 06:17:08 Consumer Terminated
PASS
ok      module1 5.009s
```