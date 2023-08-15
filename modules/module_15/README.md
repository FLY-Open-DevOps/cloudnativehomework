# 云原生训练营毕业总结

# Kubernetes的API对象

四个通用属性

## TypeMeta
GKV - Group, Kind, Version
定义在yaml文件中的apiVersion以及Kind字段字段，比如：
```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
```
Gruop是networking.istio.io
Kind是Gateway
Version是v1beta

## MetaData
* Name
    对象名称，在同一个Namespace下唯一
* Namespace
    命名空间，Namespace和Name定义了某个唯一的对象实例
* Label
    对象的可识别属性，Kubernetes API以Label作为过滤条件来查询对象(Label Seletor)
* Annotation
    作为扩展属性，面向管理员和开发人员，类似编程语言中的一些“注解”
* Finalizer
    资源锁，当其不为空的时候只能对对象做逻辑删除（更新deletetionTimestamp字段），当对象满足删除条件的时候将该字段置空，才可删除对象
* ResourceVersion
    每个对象都有一个Resource Version，起到乐观锁的作用。当请求的版本落后于当前对象记录的版本的时候，操作会被拒绝

## Spec
用户期望状态，由用户来定义

## Status
对象实际状态，由控制器更新

## 核心对象概览
分别是有：
* Node
    集群worker Node
* Namespace
    一组资源和对象的集合，常见对象都属于某个namespace之下，但是Node,PersistentVolumes这些集群级别的对象则不属于任何namespace
* Pod
    一组容器的集合，Kubernetes中资源调度的基本单位，Pod中的容器共享网络namespace，通过挂载存储卷共享存储，共享Security Context
    Pod中不同容器可以通过loopback进行访问
* 存储卷
    * Volume：定义Pod的存储卷来源
    * VolumeMounts：挂载定义好的Volume到容器内部
* 资源限制
    通过cgroup控制每个容器的CPU和内存使用
* 健康检查
    三种探针
    1. LivenessProbe
        探测应用是否处于健康状态
    2. ReadinessProbe
        Pod启动时探测应用是否就绪
    3. StartupProbe
        Pod启动时探测应用是否完成启动，当一个应用需要较长的启动时间，则需要设置StartupProbe，让探活操作的间隔变长，放置过于频繁的探活使得应用没有机会启动成功。一旦应用启动成功，该Probe会失效，让Readiness Probe来接管
* ConfigMap
    应用配置信息
* Secret
    密钥的配置信息，与ConfigMap类似，只是不用明文存储，用Base64编码的形式存储信息
* User Account && Service Account
    1. User Account
    对应用户账户，跨namespace
    2. Service Account
    运行中的程序的身份，与namespace是相关的
* Service
    应用服务的抽象，通过label为应用提供服装均衡和服务发现
* Replica Set
    控制Pod副本数
* Deployment
    控制Replica Set以及Pod的配置
* StatefulSet
    用于部署有状态应用
* Job
    用于一次性任务
* DaemonSet
    每个节点上都存在的，特殊用途的应用，比如日志采集，性能指标采集等
* CRD
    自定义资源


# Kubernetes架构

一个标准的Kubernetes集群通常由一个或以上的Master Node以及一个或以上的Worker Node组成

## Master Node

### API Server
提供Kubernetes集群管理的RESTful API，主要涵盖认证，授权以及准入三大部分的内容。

#### 认证
支持多种认证机制与插件
1. X509证书
    API Server启动时配置`--client-ca-file=`filepath`
2. 静态Token文件
    API Server启动时配置`--token-auth-file=`filepath`
    提供一个csv文件，至少包括`token, username, user id`三列，之后的列是可选的group列
3. Bootstrap Token
    动态管理的持有者令牌，以secret的形式保存在kube-system的namespace之中
4. 静态文件密码
    API Server启动时配置`--basic-auth-file=`filepath`
    提供一个csv文件，至少包括`password, username, user id`三列，之后的列是可选的group列
5. Service Account
    由kubernetes生成，挂载到容器的``路径下
6. OpenID
    OAuth 2.0 认证机制
7. WebHook令牌
    > 与企业内部的认证系统进行集成的主要形式
    * 首先需要在api server启动的时候添加以下配置：
        `--authentication-token-webhook-config-file="your config file"`指定配置文件来描述如何进行WebHook的远程访问(如配置远程服务的路径)
        `--authentication-token-webhook-cache-ttl`设定身份认证的缓存时间，默认2min
    * 开发一个用于解析api server发送的token，以及对接真正认证的服务，并将认证结果放回给api server的一个服务
    * 在`~/.kube/config`文件中指定新用户与用户对应的token，然后尝试以`kubectl --user "new user"`的形式尝试访问资源
    * 在第三方认证服务的时候需要注意控制调取真正的认证的频率，避免认证服务压力过大导致故障，常用方法是熔断和限流
8. 匿名请求
    不建议开启，可以使用`--anonymous-auth=false`禁用

#### 授权
kubernetes种常用的授权管理是RBAC模型，授权仅处理以下请求属性：
1. user,group,extra
2. API, request method, and path
3. resources && subresources
4. namespaces
5. api groups

    主要通过Role和ClusterRole来定义权限的集合，通过RoleBinding和ClusterRoleBinding来绑定特定用户与角色

#### 准入控制
> 1. 平台希望对请求对象增加一些附加属性，对原始对象做变形
> 2. 原始对象或者变形后对象是否合法
> 
> 常用场景：配额管理，控制用户在某个namepsace下面建多少个pod，service，ingress等


相关插件：
* AlwaysAdmit
* AlwaysPullingImages
* DengEscalatingExec
* ImagePolicyWebhook
* ServiceAccout
* SecurityContextDeny
* ResourceQuota：限制Pod的请求不会超过配额，需要在namespace下面创建该对象才能生效
... 更多控制器参考[官网](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#what-does-each-admission-controller-do)

准入控制插件的开发
* MutatingWebhookConfiguration：用于变形请求对象
* ValidatingWebhookConfiguration：用于校验请求对象（变形后）

开发流程：
    1. 编写一个服务来接受api server发过来的请求体，解析AdmissionReview的Request
    2. 根据业务逻辑生成AdmissionReview的Response部分，主要是UID，Allowed和Patch三个字段
    3. 将服务容器化，部署到kubernetes集群中，并配置service
    4. 根据服务暴露的api路径配置`MutatingWebhookConfiguration`或者`ValidatingWebhookConfiguration`对象（注意，这里服务一定要是https的服务，因此自签名证书的生成和绑定）

#### 限流
API Server中提供两个参数进行限流：
    * max-requests-inflight：单位时间内最大请求数
    * max-mutating-requests-inflight：单位时间内最大mutating请求数

API Priority and Fairness - APF
> 对请求进行分类，不同优先级的并发资源是隔离的，不同优先级的i元不会相互排挤，特定优先级高的请求被高优先处理
* 多等级
* 多队列

APF实现以来两个资源
* FlowSchema
    `kubectl get flowschema`
    代表一个请求的分类，并且还可以根据distinguisher进一步划分不同的Flow，flowschema name+distinguisher会唯一确定一个flow（请求）
    FS中会设置一个Priority Level(PL)，代表请求类型的优先级
    一个Priority Level可以对应多个FS，PL中维护一个QueueSet（多个queue的集合），用来缓存不能及时处理的请求

    ```bash
    $ kubectl get flowschema
    NAME                           PRIORITYLEVEL     MATCHINGPRECEDENCE   DISTINGUISHERMETHOD   AGE   MISSINGPL
    exempt                         exempt            1                    <none>                25d   False
    probes                         exempt            2                    <none>                25d   False
    system-leader-election         leader-election   100                  ByUser                25d   False
    endpoint-controller            workload-high     150                  ByUser                25d   False
    workload-leader-election       leader-election   200                  ByUser                25d   False
    system-node-high               node-high         400                  ByUser                25d   False
    system-nodes                   system            500                  ByUser                25d   False
    kube-controller-manager        workload-high     800                  ByNamespace           25d   False
    kube-scheduler                 workload-high     800                  ByNamespace           25d   False
    kube-system-service-accounts   workload-high     900                  ByNamespace           25d   False
    service-accounts               workload-low      9000                 ByUser                25d   False
    global-default                 global-default    9900                 ByUser                25d   False
    catch-all                      catch-all         10000                ByUser                25d   False
    ```
* PriorityLevelConfiguration
    优先级配置
    每个优先级配置中定义该优先级的队列数目，允许的并发请求数，每个flowschema+distinguisher的请求最多能被enqueue多少个队列，队列中的元素个数等信息
    ```bash
    $ kubectl get prioritylevelconfiguration
    NAME              TYPE      ASSUREDCONCURRENCYSHARES   QUEUES   HANDSIZE   QUEUELENGTHLIMIT   AGE
    catch-all         Limited   5                          <none>   <none>     <none>             25d
    exempt            Exempt    <none>                     <none>   <none>     <none>             25d
    global-default    Limited   20                         128      6          50                 25d
    leader-election   Limited   10                         16       4          50                 25d
    node-high         Limited   40                         64       6          50                 25d
    system            Limited   30                         64       6          50                 25d
    workload-high     Limited   40                         128      6          50                 25d
    workload-low      Limited   100                        128      6          50                 25d
    ```
调试工具：
* `kubectl get --raw /debug/api_priority_and_fairness/dump_requests`
    显示出当前队列的堆积情况


#### apimachinery 
TODO


### ETCD
一个基于Raft协议的分布式KV存储，提供KV增删改查，监听，key过期和续约，以及leader选举等特性；
在Kubernetes中，负责记录集群中的各种资源的信息，例如Deployment，Service，Pod等信息，并且仅能通过API Server进行ETCD的相关操作（因为API Server可以提供数据缓存来减少对ETCD的直接访问，减轻ETCD的压力）；

### Controller Manager
Controller是用于管理每种对象的状态，确保集群中对象的真实状态和用户定义的期望状态保持一致（如Deployment种的Replica数目）；

Controller Manager则是多个Controller的组合，每个Controller一直处于一个循环之中不断去监听它所负责的对象，当对象发生变化的时候完成配置，如果配置失败则不断重试，以能达到用户定义的期望状态

#### Controller的工作原理
Controller由两个核心的组件，分别是Informer和Lister
* Lister
    Controller对于Kubernetes中的对象的一个缓存，通过Lister可以在本地通过key直接找到一个对象的当前状态，而不需要每次查询都需要跟API Server进行一次通信，减轻API Server的压力

* Informer
    被监听的对象的`增删改`操作都会以时间的形式通知到Informer，Informer将对象对应的key信息放入FIFO的work queue，等待worker协程或线程不断地从queue中取出，然后对对象进行相应的操作

### Scheduler
是一个特殊的Controller，职责是监控集群内所有没有进行调度的Pod，根据所有Worker Node的健康状态与资源使用情况，为Pod选择最佳的Node，完成调度

调度分为Predict（过滤资源无法满足要求的Node），Priority（节点打分）和Bind（绑定节点）三个阶段，

## Worker Node

### Kubelet
主要有一下职责：
* 获取Pod list，按序启动或者停止Pod
* 汇报Node的资源信息以及健康状态
* Pod健康检查以及状态汇报

### Kube Proxy
监控集群中和用户发布的Service，并完成Load Balance的配置

