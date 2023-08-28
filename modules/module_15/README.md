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

    Pod的生命周期
    Pod的状态由`status.phase`和`status.conditions`计算出来的

    如何保证Pod高可用：
    * 设置合理的resource limit（内存和emptyDir），放置资源不足被evict

    Pod的QOS（Quality of Service）
    * Graranteed：一定需要指定资源才可以运行（仅设置resources.request或者resource.request = resource.limit）
    * Burstable：仅需一定资源即可启动，但是设置最大可抢占资源数量，适用绝大多数场景（resource.request < resource.limit）
    * Besteffort：不去设置resources，放任资源竞争

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
    ReadinessGates
        自定义的就绪条件，需要使用额外的控制器。以上所有探针都就绪了，它还没就绪时说明Pod依旧还没有就绪

    两个hook
    * PostStart
        应用成功启动（成功进入entrypoint）后，做的一些事情，无法保证与entrypoint的执行先后顺序，但是poststart完成，container就会就绪
    * PreStop
        container被删除的时候可以插入的行为（如优雅终止的行为）

    terminationGracePeriodSeconds
    代表了整个grace period的时间长度，顺序是：prestop -> kill -SIGTERM -> kill -SIGKILL
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

    - 负载均衡
        监控集群中和用户发布的Service，并完成Load Balance的配置

        网络包
        - 三层：IP相关，IP header
        - 四层：端口相关，TCP header
        - 七层：应用协议相关，如HTTP header

        负载均衡
        - 集中式
            主要是外部流量先通过集中式LB，再进入到集群中
        - 进程内
            语言相关，强耦合
        - 独立进程

        相关技术
        - NAT
            负载均衡器通过修改源和目标地址来控制数据包转发行为
        - 新建TCP连接
            和NAT类似，不同的点是它会断掉源端的链接，再与对端建立新的链接
        - 链路层负责均衡
            这种情况下负责均衡器和上游服务器要在同一个IP地址下，负责均衡器收到请求之后直接修改链路层MAC地址直接发送到对应的服务去
        - 隧道技术
            在IP头外层增加额外的IP包头然后转发给上游，类似overlay

    - Service相关对象
        - Endpoints

            Endpoint controller监听service对象创建后，根据label selector，获取到对应的Pod IP的集合，记录到addresses属性中

            1. 如果Pod not ready，加入的是`subnets.notReadyAddresses`，ready的Pod加入`subnets.addresses`
            2. 如果配置了PublishNotReadyAddress为true，则无论是否ready都加入address中

        - Endpoint slice

            高版本中kubernetes的一个针对endpoint做的性能优化

            如果一个service背后的pod非常多（上千个级别），那么每次pod就绪或者生存状态发生变动的时候就要将整个很大的endpoint配置文件推送给kube-proxy，并且如果抖动情况发生频繁的时候，就会造成性能问题
            Endpoint Slice将全部的Endpoint切分为若干个切片，每次变动的时候仅需要推送包含该变动的切片到kube-proxy即可，优化了性能

        > 如何为集群外面的一组服务配置service?
        >
        > 定义一个无label selector的service，然后认为创建endpoint/endpoint slice，在subset中填写集群外的IP或者域名

        - Service
            1. Label Selector
            2. Port转换

            类型：
            - clusterIP
            - nodePort
            - loadBalancer
            - Headless Service
            - ExternalName Service

            Service Topology
            kubernetes通过提供标签来表示节点的物理区域位置
            Service可以引入topologyKey属性来进行流量控制
            > TODO: Service的`spec.topologyKey`已经在1.22被停用，在高版本需要怎么做

        
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

> 本小段主要涉及CRD和对应controller的开发（Operator模式）

Group的定义：[kubernetes/pkg/apis/core/register.go](https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/register.go)
List
单一对象的数据结构：详见[MetaData](#MetaData)
想要去自定义对象的时候，可以使用[Code Generator](https://github.com/kubernetes/code-generator)，通过定义对象，以及标注特定的tag，通过code generator即可为对象生成kubernetes对象中的一些特定的方法，如deepCopy等

etcd storage的实现：[kubernetes/pkg/registry/core/configmap/storage/storage.go](https://github.com/kubernetes/kubernetes/blob/master/pkg/registry/core/configmap/storage/storage.go)

subresource
内嵌在kubernets对象中，有独立的操作逻辑的属性集合，如pod status，pod status需要频繁更新，此时在podStatusStrategy里面可以定义更新pod status的时候用旧的pod spec覆盖掉新的pod spec保证更新pod status的时候不会改动到pod spec，避免了reversion的影响




### ETCD
一个基于Raft协议的分布式KV存储，提供KV增删改查，监听，key过期和续约，以及leader选举等特性；
在Kubernetes中，负责记录集群中的各种资源的信息，例如Deployment，Service，Pod等信息，并且仅能通过API Server进行ETCD的相关操作（因为API Server可以提供数据缓存来减少对ETCD的直接访问，减轻ETCD的压力）；

### Controller Manager
Controller是用于管理每种对象的状态，确保集群中对象的真实状态和用户定义的期望状态保持一致（如Deployment种的Replica数目）；

Controller Manager则是多个Controller的组合，每个Controller一直处于一个循环之中不断去监听它所负责的对象，当对象发生变化的时候完成配置，如果配置失败则不断重试，以能达到用户定义的期望状态

Kubernetes中默认开启的通用的一些Controller，如：Deployment Controller, Job Controller, Service Controller等

> Cloud Controller Manager，这些controller往往跟云厂商深度集成，因此被分离作为独立的Controller manager, 例如定制的IngressController，Service Controller等

各个controller的启动源码：[kubernetes/cmd/kube-controller-manager/app/core.go](https://github.com/kubernetes/kubernetes/blob/master/cmd/kube-controller-manager/app/core.go)


#### Controller的工作原理

Controller由两个核心的组件，分别是Informer和Lister
* Lister
    Controller对于Kubernetes中的对象的一个缓存，通过Lister可以在本地通过key直接找到一个对象的当前状态，而不需要每次查询都需要跟API Server进行一次通信，减轻API Server的压力

* Informer
    被监听的对象的`增删改`操作都会以时间的形式通知到Informer，Informer将对象对应的key信息放入FIFO的work queue，等待worker协程或线程不断地从queue中取出，然后对对象进行相应的操作


确保scheduler和controller的HA
使用一个controller持续watch某个configmap和endpoint（kubernetes提供的leader election的类库）的annotation信息，leader会把自己的信息更新到endpoint的annotation上，并在一段时间内要回来renew，Lease对象


### Scheduler
是一个特殊的Controller，职责是监控集群内所有没有进行调度的Pod，根据所有Worker Node的健康状态与资源使用情况，为Pod选择最佳的Node，完成调度（更新Pod的NodeName字段）

调度分为Predicate（过滤资源无法满足要求的Node），Priority（节点打分）和Bind（绑定节点）三个阶段

#### Predicate
Predicate根据一系列的策略（一些列的Predicate Plugins）来过滤资源无法满足的Node，策略包括端口冲突，计算资源是否满足（CPU,GPU,内存），NodeSelector是否匹配，亲和性策略，是否能容忍污点等，也可以自己定义策略

在进入每一个策略插件计算完成后就过滤一部分节点，最后只剩下满足调度条件的节点列表

#### Priority
Priority根据一系列的策略（一些列的Priority Plugins）来为过滤后的每一个Node打分，策略包括Pod尽量分布在不同节点，亲和性，优先调度到请求资源少的节点，平衡各个节点资源使用等等

#### Affinity

##### NodeAffinity
基于Label Selector去过滤不符合条件的Node，有nodeAffinity和nodeAntiAffinity

主要有两种：
* requiredDuringSchedulingIgnoredDuringExecution
    硬亲和，一定要满足需求才会调度
    ```yaml
    spec:
    affinity:
        nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
            - key: disktype
                operator: In
                values:
                - ssd 
    ```
* preferredDuringSchedulingIgnoredDuringExecution
    软亲和，不满足的时候也可以作为备选节点，只是调度的权重可能会偏低
    ```yaml
    spec:
    affinity:
        nodeAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 1
            preference:
            matchExpressions:
            - key: disktype
                operator: In
                values:
                - ssd  
    ```
##### PodAffinity
基于Label Selector去查看如果Node中是否含有符合条件的Pod，有podAffinity和podAntiAffinity
* requiredDuringSchedulingIgnoredDuringExecution
    硬亲和，一定要满足需求才会调度
    ```yaml
    spec:
    affinity:
        podAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
            - key: disktype
                operator: In
                values:
                - ssd 
            topologyKey: kubernetes.io/hostname
    ```
    > topologyKey 是Node的一个label，代表可用区，上面的例子表示要跟满足标签的Pod放置在`kubernetes.io/hostname`这一个可用区内

* preferredDuringSchedulingIgnoredDuringExecution
    软亲和，不满足的时候也可以作为备选节点，只是调度的权重可能会偏低
    ```yaml
    spec:
    affinity:
        podAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 1
            preference:
            matchExpressions:
            - key: disktype
                operator: In
                values:
                - ssd 
            topologyKey: kubernetes.io/hostname
    ```

#### Taints && Tolerations
用于保证Pod不会被调度到不适合的Node上，Taint作用于Node，Toleration作用于Pod

##### Taints

```bash
$ kubectl taint nodes node1 key1=value1:NoSchedule
```

Taint类型
* NoSchedule：新的Pod不应该调度到该Node，不影响症状运行的Pod
* PreferNoSchedule：新的Pod尽量不要调度到该Node
* NoExecute: 新的Pod不应该调度到该Node，且驱逐已经运行的Pod

##### Tolerations
Pod可以设置是否容忍某些Taint，如果满足，则可以调度或者不被驱逐

```yaml
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  tolerations:
  - key: "example-key"
    operator: "Exists"
    effect: "NoSchedule"
```

#### 调度优先级
为Pod区分优先级，保证优先级高的Pod优先调度，或者在资源不足的时候区组低优先级的Pod从而获得调度资源

* api-server配置 `--feature-gates=PodPriority=true` 和 `--runtime-config=scheduling.k8s.io/v1alpha1=true`
* kube-scheduler配置 `--feature-gates=PodPriority=true`

Priority Class
定义Pod的优先级，定义该优先级的value，是否全局默认（globalDefault），以及是否可以被抢占（preemptionPolicy）
```yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000000
preemptionPolicy: Never
globalDefault: false
description: "This priority class should be used for XYZ service pods only."
```

在Pod中指定priority
```yaml
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  priorityClassName: high-priority
```

#### 自定义调度器
默认使用DefaultScheduler进行调度，但如果默认调度器不满足需求，也可以自定义调度器，并且在PodSpec中指定期望的自定义调度器名称
```yaml
spec:
  schedulerName: my-scheduler
  containers:
  - name: pod-with-default-annotation-container
    image: registry.k8s.io/pause:2.0
```


## Worker Node

### Kubelet
每个worker节点都有运行一个kubelet服务进程，默认端口10250

主要有一下职责：
* 获取Pod list，按序启动或者停止Pod
* 汇报Node的资源信息以及健康状态
* Pod健康检查以及状态汇报

架构：
* Kubelet API
    1. 探活
    2. 业务指标上报
* Managers
    各个不同职责的控制器，如探活，OOM监控，磁盘空间管理，cAdvisor(基于cgroups技术获取节点上运行的资源情况)，syncLoop(接受Pod变化通知)，PodWorker(更新Pod)

    > syncLoop接受来自API Server的Pod更新通知，将时间存放到UpdatePodOptions中（里面是多个queue）;
    > 
    > worker从queue中获取Pod事件的变更内容清单，然后针对每一个Pod进行syncPod的操作;
    > 
    > 调用CRI接口真正对Pod进行创建或更新；
    >
    > 通过PLEG组件上报Pod状态到syncLoop，再上报到api server
* CRI

#### Pod 管理
* 获取Pod列表
    1. 文件（静态Pod，集群启动的时候默认创建的Pod）
    2. HTTP endpoint，启动参数配置`--manifest-url`，将清单放在url中，效果类似文件形式
    3. API Server

Kuberlet在启动容器进程的时候，是启动多个容器进程（即使PodSpec里面只是申明了一个容器）
* pause
    是一个sandbox进程，比所有业务容器都先被拉起，用于挂载network namespace，相当于每个业务容器的底座，业务容器发生问题也不会影响到Pod的网络配置，业务容器重启也无需重新配置网络。
    pause启动之后，containerd会调用cni插件为Pod配置网络，配置完之后返回给运行时，运行时上报给kubelet，此时pod就有了IP
    要查看pause，需要在worker节点中通过ctr进行查看
    ```bash
    $ ctr -n k8s.io c list | grep pause
    0fd6a2572faf674484a1edca18b122810af9464057357d2d11ae3b37d61ae754    registry.aliyuncs.com/google_containers/pause:3.6             io.containerd.runtime.v1.linux    
    1686d6b758b60cd7a55077706591bcb01c55e9e1ee09b430627f33a96c77aa66    registry.aliyuncs.com/google_containers/pause:3.6             io.containerd.runtime.v1.linux    
    1c985ccc1b147746e8969cfa63540b248a7d029516f1bf2b9fc7e2f2130187a8    registry.aliyuncs.com/google_containers/pause:3.6             io.containerd.runtime.v1.linux    
    301053269db503943fa30e07700353bfbdcc6c31db86546242b66953b2cf7204    registry.aliyuncs.com/google_containers/pause:3.6             io.containerd.runtime.v1.linux
    ```
Kubelet启动Pod的流程：
    TODO
    总体是CSI -> CRI -> CNI

#### CRI
container runtime interface，kubernetes定义的一组GRPC的服务
包括两类服务：
* 镜像服务
    遵循OCI的Image Specification
* 运行时服务
    遵循OCI的Runtime Specification
    1. CRI runtime：与kubelet交互，如containerd，CRI-O等
    2. OCI runtime：与具体的容器运行时交互，如runc，kata等

#### CNI
container network interface，kubernetes通过提供一个轻量级的通用的容器网络接口CNI，专门用于设置和删除容器的网络连通性，容器运行时通过CNI调用网络插件来完成容器的网络设置
CNI的配置是通过直接调用二进制文件来执行的

kubernetes网络模型的设计基本原则：
1. 所有Pod无需NAT就能互相访问
2. 所有节点无需NAT就能互相访问
3. 容器内看见的IP地址和外部组件看到容器的IP是一致的

##### CNI插件分类
* IPAM: 分配IP地址
* 主插件：网卡设置
    * bridge
    * ipvlan，和bridge一样，用于打通主机和容器的网络
    * loopback
    * 还有其它社区提供的，如Calico，Cilium等
* Meta：附加功能
    * portmap：设置主机与容器端口映射
    * bandwidth：限流
    * firewall：利用iptables或者firewalld为容器设置防火墙规则

默认CNI配置路径`/etc/cni/net.d`, 一个CNI配置的例子：
```json
{
    "cniVersion": "0.3.1",
    "name": "crio",
    "type": "bridge",
    "bridge": "cni0",
    "isGateway": true,
    "ipMasq": true,
    "hairpinMode": true,
    "ipam": {
        "type": "host-local",
        "routes": [
            { "dst": "0.0.0.0/0" },
            { "dst": "1100:200::1/24" }
        ],
        "ranges": [
            [{ "subnet": "10.85.0.0/16" }],
            [{ "subnet": "1100:200::/24" }]
        ]
    }
}
```
配置文件里面的type指明了CNI插件的名称，比如bridge，calico等

##### 常见插件
* Flannel
    基于VxLAN，在原始数据报封装一层包头（overlay）去做转发，简单易用，但是效率不高，且流量跟踪比较困难
* Calico
    性能好，灵活，支持网络策略
    DaemonSet的形式运行，主要组件有：
    1. felix agent：配置防火墙规则
    2. BIRD agent：BIRD是一个路由交换软件，将主机模拟为一个路由器，主机之间基于BGP协议交换路由信息
    3. confd agent：用来进行配置推送

    calico的模式：
    * VXLan
    * BGP

    初始化：
    由init container通过mount host path的形式将calico相关的二进制文件配置到Node中

    所在的api group：`crd.projectcalico.org`

    IP分配相关CRD
    * IPPOOLS
        定义CIDR
    * IPAM Block
        定义每台主机预分配的IP段，以及记录了哪个IP分配到了哪个Pod的信息，以及未分配的IP
        ```bash
        $ kubectl get ipamblock
        ```
    * IPAM Handler
        记录了IP分配的详细信息

##### 数据包流转
[reference](https://dramasamy.medium.com/life-of-a-packet-in-kubernetes-part-2-a07f5bf0ff14)

#### CSI
> 插件管理的形式
> * in-tree
>     不再接受新的插件
> * out-of-tree
>     * FlexVolume
>         需要root权限
>     * CSI
>         主流

通过RPC调用形式与存储驱动进行交互

结构：
* 存储控制器
    存储适配器
    * provisioner： 创建volume
    * attacher：挂载volume
    CSI驱动

* 工作节点
    存储代理
    * 节点驱动注册器
    * CSI代理


常见CSI：
* 临时存储：
    * emptyDir
        与Pod生命周期强绑定，Pod一旦重建，则会被抹除
* 半持久化：
    * host path
        永久持久化到工作节点的特定位置中

    > host path可以以PV/PVC的形式暴露给终端用户
    > 
    > 1. 手工创建host path对应的StorageClass为manual的PV
    >
    > 2. 终端用户通过PVC去使用该PV(PV中的StorageClass需要与PV保持一致)

* 永久存储相关的资源
    1. StorageClass
        定义provisioner以及mount的相关参数
    2. Volume
    3. Persistent Volume Claim
        由用户创建，代表用户对存储的需求声明
        定义StorageClass，存储大小，访问模式（读写相关）等
    4. Persistent Volume
        一般由集群管理员提前创建，或者根据PVC需求动态创建，代表后端的真实存储空间
    
##### Rook
云原生环境下的开源分布式存储编排系统

架构：
* Rook Operator
    负责拉去Rook其它组件，启动控制平面
* Rook Discover
    负责发现物理存储空间
* Provisioner
* Rook Agent
    DaemonSet，处理所有存储操作，如mount存储卷到容器等





### Kube Proxy

每个worker node上都会运行一个kube proxy服务，负责监听api server中的service和endpoint的变化情况，通过iptables等来为服务配置负载均衡（仅TCP和UDP）

形式：直接运行在物理机 / static pod / DaemonSet

实现方式：
- userspace
- iptables
    完全基于iptables规则来实现service的负载均衡
    用户态配置NAT表的规则，当要访问ClusterIP的地址的时候，数据包的转发到真实的后端Pod中
- ipvs
    解决iptables的性能问题，采用增量更新的形式保证service在更新的时候连接不会断开，一般情况推荐使用
- winuserspace
    windows环境下的实现



##### Netfilter框架
Linux内核处理接受到的数据包的流转的框架，主要作用于网络层
> 一个数据包通常关注的5元组：协议，源地址，源端口，目标地址，目标端口

四表五链
- 四表： raw，mangle，nat，filter
- 五链：prerouting，input，forward，output，postrouting，这些都暴露了HOOK供用户配置特定规则
    一般是prerouting和output两个HOOK对目标地址进行修改

外部数据包进入Linux的流程
1. 网卡驱动接受到数据包，对CPU发起硬件中断，将数据包通过直接访问内存的形式复制到Kernal的某块内存空间中
2. CPU发起软中断，调用softIRQ handler，将接收到的数据包构造城SK Buffer，SKB分为两部分：Header（五元组）+ Data，并将SKB交付到Netfilter框架
3. Netfilter去用户态读取定义的iptables的规则来处理数据包

#### Kube Proxy工作原理

##### iptables模式为例

Kube Proxy调用iptables的命令生成转发规则（规则大致是如果数据包的目标ip是clusterIP就dnat到特定的某个podIP中）
下面是一个具体实例，比如我们发布了一个三个Pod的服务，如下：
```bash
$ kubectl get pods -n cloudnative -owide
NAME                                     READY   STATUS    RESTARTS   AGE     IP            NODE       NOMINATED NODE   READINESS GATES
http-server-demo-dev-6c7678cdcb-kdrg7    1/1     Running   0          7m12s   172.17.0.14   minikube   <none>           <none>
http-server-demo-dev-6c7678cdcb-vkdvk    1/1     Running   0          7m12s   172.17.0.12   minikube   <none>           <none>
http-server-demo-dev-6c7678cdcb-wj6b2    1/1     Running   0          7m12s   172.17.0.13   minikube   <none>           <none>

$ kubectl get svc -n cloudnative
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
http-server-demo-dev    ClusterIP   10.98.140.123   <none>        80/TCP    89s
```

通过iptables-save命令查看nat表下的规则，找到和CLUSTER-IP相关的规则
```bash
# 从prerouting开始，所有请求都强制需要经过KUBE-SERVICE链
-A PREROUTING -m comment --comment "kubernetes service portals" -j KUBE-SERVICES

# 从KUBE-SERVICE链中获得的数据包，对10.98.140.123为目标IP且端口为80时，执行规则KUBE-SVC-7MAVMYXEWSWTZO74
-A KUBE-SERVICES -d 10.98.140.123/32 -p tcp -m comment --comment "cloudnative/http-server-demo-dev cluster IP" -m tcp --dport 80 -j KUBE-SVC-7MAVMYXEWSWTZO74

# 以33%的概率命中第一条规则，转发到规则KUBE-SEP-I3FOTHZ2ZFHDEZEY
-A KUBE-SVC-7MAVMYXEWSWTZO74 -m comment --comment "cloudnative/http-server-demo-dev -> 172.17.0.12:80" -m statistic --mode random --probability 0.33333333349 -j KUBE-SEP-I3FOTHZ2ZFHDEZEY
# 以50%的概率命中第一条规则，转发到规则KUBE-SEP-VNYQZSEJHSR6FRJ3
-A KUBE-SVC-7MAVMYXEWSWTZO74 -m comment --comment "cloudnative/http-server-demo-dev -> 172.17.0.13:80" -m statistic --mode random --probability 0.50000000000 -j KUBE-SEP-VNYQZSEJHSR6FRJ3
# 以上规则都不命中，作为兜底规则，100%转发到规则KUBE-SEP-A7FT4NOFGEXI2SSB
-A KUBE-SVC-7MAVMYXEWSWTZO74 -m comment --comment "cloudnative/http-server-demo-dev -> 172.17.0.14:80" -j KUBE-SEP-A7FT4NOFGEXI2SSB

# dnat到Pod 172.17.0.13:80
-A KUBE-SEP-VNYQZSEJHSR6FRJ3 -p tcp -m comment --comment "cloudnative/http-server-demo-dev" -m tcp -j DNAT --to-destination 172.17.0.13:80
# dnat到Pod 172.17.0.14:80
-A KUBE-SEP-A7FT4NOFGEXI2SSB -p tcp -m comment --comment "cloudnative/http-server-demo-dev" -m tcp -j DNAT --to-destination 172.17.0.14:80
# dnat到Pod 172.17.0.12:80
-A KUBE-SEP-I3FOTHZ2ZFHDEZEY -p tcp -m comment --comment "cloudnative/http-server-demo-dev" -m tcp -j DNAT --to-destination 172.17.0.12:80
```
iptables的规则都是从上到下执行的，从上面的规则可以看出iptables的LB逻辑
存在问题：
- 当Pod的数量很大的时候，数据包需要hit非常多的规则，因此数据包的转发效率不高
- iptables要刷新规则的时候，kube-proxy无法做增量检查，只能把规则清除掉从头开始写，因此刷新规则的资源消耗非常高
因此在iptables的模式下，整个集群的规模不可以太大，service不能太多，service后面的pod也不能太多

iptables模式下的service的cluster ip可以是一个不用绑定在任何设备上的虚拟ip，只是用来匹配PREROUTING的规则而已，不经过路由判定

##### ipvs工作模式
ipvs是LVS(linux virtual servivce)一部分

ipvs是在INPUT和OUTPUT链上订制规则，因此此时service上的clusterIP必须是本机的一个有效IP（绑在当前节点的一个dummy设备上）

从iptables模式切换为ipvs模式：
```bash
$ kubectl edit cm kube-proxy -n kube-system
# set mode: "ipvs"
# then remove the kube-proxy container to let it rebuild
$ iptables --flush
```

查询使用ipvsadm，用法和iptables大致相同


### DNS服务
Kubernetes中负责集群内的DNS服务的是coreDNS，coreDNS负责监听Service和EndPoint的变化并配置DNS
```bash
$ kubectl get pods -n kube-system
NAME                               READY   STATUS    RESTARTS       AGE
coredns-565d847f94-mj9wk           1/1     Running   9 (21h ago)    28d

$ kubectl get service -n kube-system
NAME       TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                  AGE
kube-dns   ClusterIP   10.96.0.10   <none>        53/UDP,53/TCP,9153/TCP   28d

$ kubectl get deploy -n kube-system
NAME      READY   UP-TO-DATE   AVAILABLE   AGE
coredns   1/1     1            1           28d
```

业务Pod进行域名解析的时候，就是从coreDNS中查询域名对应的IP的，业务pod中配置的dns server就是core dns
```bash
$ kubectl exec loki-0 -- cat /etc/resolv.conf
nameserver 10.96.0.10
search default.svc.cluster.local svc.cluster.local cluster.local
options ndots:5

$ kubectl get service -n kube-system
NAME       TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                  AGE
kube-dns   ClusterIP   10.96.0.10   <none>        53/UDP,53/TCP,9153/TCP   28d
```
可以看到业务pod中的nameserver配置的就是kube-dns的服务地址

服务分类
- 一般service
    ClusterIP，nodePort，LoadBalancer类型的service都有api server分配的集群IP，coreDNS会为这些service创建`FQDN : IP`的A记录，PTR记录，以及为端口创建SRV记录
    > A记录：从域名查IP
    >
    > PTR记录：从IP查域名
    >
    > SRV记录：端口相关的记录

    kubernetes中FQDN的格式: `svcname.namespace.svc.clusterdomain`
- Headless service
    service中指定`spec.clusterIP:None`的时候该服务为无头服务
    此时，api server不会为该service分配一个clusterIP，coreDNS为此类service创建多条A记录，每条A记录均是`FQDN : 其中一个Pod的IP`

- External Service
    此类service用来引用一个已经存在的域名，coreDNS为其创建一个CNAME记录指向目标域名

域名解析流程：
```bash
$ kubectl exec loki-0 -- cat /etc/resolv.conf
nameserver 10.96.0.10
search default.svc.cluster.local svc.cluster.local cluster.local
options ndots:5
```

`options ndots:5`表示当域名中的点的数量在`[0:4)`以内的时候，会尝试用`search`中的内容进行填充去，尝试找到一个能成功解析的域名
比如`curl test`，就会尝试为test这个域名添加后缀`default.svc.cluster.local`，然后尝试去解析`test.default.svc.cluster.local`这个域名

在Pod中的相关配置
```yaml
spec:
  dnsPolicy: ClusterFirst #
  enableServiceLinks: true # true的时候Service的一些信息会以环境变量的形式注入POD中，当集群规模较大的时候建议关闭，否则有可能因为Service信息太多导致POD启动不成功
```

可以为Pod指定`/etc/resolv.conf`的内容，也就是DNS的解析策略
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
    - name: my-container
      image: my-image:tag
  dnsPolicy: None
  dnsConfig:
    nameservers:
      - 8.8.8.8
    searches:
      - my-domain.local  # Specify search domains here
    options:
      - name: ndots
        value: "5"  # Set the value for the ndots option
```

### Ingress
Service的负载均衡主要工作在L4，Ingress的负载均衡主要工作在L7，本质上就是一个反向代理的软件，如Nginx，Envoy等

Ingress职责：
- 在Ingress层集中管理证书，以及集群外部的请求的TLS termination工作是在这一层进行的
- 分析用户的request url的path和request header中的一些属性，决定这个请求要被转发到Ingress背后的哪个service中

Ingress Controller
- 负责生成相应的反向代理软件的相关配置
- 一个集群中可能运行着多个Ingress Controller，在创建Ingress对象的时候需要指定`spec.ingressClassName`为Ingress controllr的名字来选择生成不同软件的转发配置

# 生产集群管理
## OS选择
> TODO

## 节点资源管理

- 状态上报
Kubelet通过[Lease](https://kubernetes.io/docs/concepts/architecture/leases/)对象向API server汇报自身的健康状态，如果超过nodeLeasesDurationSeconds（默认40s）没有更新状态，则节点会被判定为不健康

查看节点lease信息：
```bash
$ kubectl get lease -n kube-node-lease minikube -oyaml
```
```yaml
apiVersion: coordination.k8s.io/v1
kind: Lease
metadata:
  creationTimestamp: "2023-07-20T13:20:15Z"
  name: minikube
  namespace: kube-node-lease
  ownerReferences:
  - apiVersion: v1
    kind: Node
    name: minikube
    uid: f36c7602-3b2f-4384-ba96-1beda127c62d
  resourceVersion: "280880"
  uid: 7caf27f1-5f00-4684-a5c0-0fec005c9e85
spec:
  holderIdentity: minikube
  leaseDurationSeconds: 40
  renewTime: "2023-08-18T09:38:57.963829Z"
```

查看节点的健康信息：
```bash
$ kubectl get nodes -oyaml
```
关注以下内容：
```yaml
status:
  conditions:
  - lastHeartbeatTime: "2023-08-18T09:30:39Z"
    lastTransitionTime: "2023-08-16T14:32:01Z"
    message: kubelet has sufficient memory available
    reason: KubeletHasSufficientMemory
    status: "False"
    type: MemoryPressure
  - lastHeartbeatTime: "2023-08-18T09:30:39Z"
    lastTransitionTime: "2023-08-16T14:32:01Z"
    message: kubelet has no disk pressure
    reason: KubeletHasNoDiskPressure
    status: "False"
    type: DiskPressure
  - lastHeartbeatTime: "2023-08-18T09:30:39Z"
    lastTransitionTime: "2023-08-16T14:32:01Z"
    message: kubelet has sufficient PID available
    reason: KubeletHasSufficientPID
    status: "False"
    type: PIDPressure
  - lastHeartbeatTime: "2023-08-18T09:30:39Z"
    lastTransitionTime: "2023-08-16T14:32:01Z"
    message: kubelet is posting ready status
    reason: KubeletReady
    status: "True"
    type: Ready
```

- 资源预留
    为Kubernetes的基础服务预留CPU，内存，PID等资源，保证集群正常工作，相关参数`SystemReserved`，`KubeReserved`等
    查询节点的资源预留情况
    ```bash
    $ kubectl get nodes -oyaml
    ```
    查看以下内容：
    ```yaml
    status:
        allocatable:
            cpu: "6"
            ephemeral-storage: 80698128Ki
            hugepages-1Gi: "0"
            hugepages-2Mi: "0"
            memory: 12221572Ki
            pods: "110"
        capacity:
            cpu: "6"
            ephemeral-storage: 80698128Ki
            hugepages-1Gi: "0"
            hugepages-2Mi: "0"
            memory: 12221572Ki
            pods: "110"
    ```
    capacity指整个节点拥有的计算资源
    allocatable指节点可以分配给其它作业的计算资源

    节点Capacify包括以下内容：
    - kube-reserved
    - system=reserved
    - eviction-threshold
        需要启动驱逐行为的资源消耗阈值
    - allocatable

    节点磁盘管理
    > TODO

- 驱逐管理
    当节点资源紧张的时候，kubelet回去驱逐一些Pod，但是不是删除Pod，而是只是终止了Pod中的容器进程，留下驱逐的痕迹

    驱逐策略
    - evictionSoft：满足条件，等待一个宽限期，如果宽限期到了还是满足条件，则采取驱逐行为，并且给pod优雅终止的机会
    - evictionHard: 满足条件，直接驱逐，直接使用杀死Pod中的容器
    主要配置在kubelet启动的时候的`--config`中的配置文件决定，默认路径是`/var/lib/kubelet/config.yaml`
    ```yaml
    evictionHard:
        imagefs.available: 0%
        nodefs.available: 0%
        nodefs.inodesFree: 0%
    ```

## 配置高可用集群
- KubeSpary
    > TODO

- ClusterAutoScaler
    > TODO


# 生产化运维

## 镜像相关

### 镜像仓库
Harbor搭建进行仓库
> TODO

### 镜像anquan
> TODO

## DevOps

一个应用ready，需要包括以下内容：
- function ready，业务功能完善
- prodution ready
    - 通过负载与压力测试
    - 完成用户手册
    - 完成管理手册，可以按照管理手册部署与升级服务
    - 监控
        - 组件健康检查
        - 性能指标（Metrics）
        - 基于性能指标定义alert rule
        - 定期测试某功能


## CICD

### Github Action
基于Github的Action构建流水线

[文档](https://docs.github.com/en/actions)

### Jenkins
基于脚本配置流水线
有以下问题
- 复用率不足
- 代码调试较为困难

### Tekton
声明式API的流水线，[文档](https://tekton.dev/docs/)

核心组件：
- Pipeline
    每个Pipeline由一个或者数个Task组成
- Task
    Kubernetes为每一个Task创建一个Pod，一个Task由多个Step组成，每个Step是Pod中的一个容器

> pipeline和task对象可以接受git registry, PR等资源作为输入，可以将Image，Kubernetes Cluster，Storage，CloudEvent作为输出

EventListener
核心属性是interceptors拦截器，可以监听多种事件类型，比如GitLab的Push事件。当EventListener对象被创建以后，Tekton的controller会为该EventListener创建Kubernetes的Pod和Service，启动一个http服务监听Push事件。当用户在GitLab中设置webhook填写EventListener服务地址后，针对该项目的Push操作都会被EventListenner捕获


### Argo CD
- 用于Kubernetes的声明式GitOps的连续交付工具
- 实现是基于Kubernetes的controller。该控制器连续监视症状运行中的aoo，并将当前的活动状态和所需的目标状态（Git repo中指定）进行比较，当活动状态偏离目标状态的已部署app会被标记为OutOfSync
- 报告可视化差异，提供自动或者手动将实时状态同步回所需要的目标状态
* 对Git repo中对目标状态做的修改都可以自动应用并反映到所指定的目标环境宏

## 日志
Grafana Loki

架构：
- Loki
    日志处理与持久化
    - Distributor
        - 处理客户端谢日的日志
        - 接收到日志后分批发送给多个Ingester
    - Ingester
        日志数据持久化（DynamoBD，S3等）
        保证日志顺序
    - Querier
        基于LogQL
- Promtail
    Pod日志搜集，并为日志打上相关Pod的label
- Grafana
    展示与查询

## 监控
Prometheus

每个节点的Kubelet（继承了cAdvisor）会收集当前节点host上所有信息，包括cpu，内存，磁盘的等，Prometheus会pull这些信息，并给每个Node打上标签来区分不同的Node的信息

Kubernetes的control panel，包括各种controller都原生地暴露了Prometheus格式的metrics

- 架构
    - Prometheus Server
        - Retriveval：采集指标
        - TSDB：存储指标
            - 特性
                纵向写：每一次数据采集，都会汇报周期内的所有指标数据
                横向读：按特定的序列读取，单调时间序列 = Metrics Name + 唯一的labels
            - 存储机制
                - 内存
                - 磁盘
                - WAL，write ahead log
            - 索引
                - 每个时间序列都有一个唯一ID
                - 索引模块维护了一个label和ID的map关系
                - 通过K路轮询完成时间序列的高效查询
            
        - Http server：查询与展示数据
    - PushGateway
        对于job程序可以将指标统一推送到pushgateway，再由Retriveval统一采集
    - AlertManager
        配置告警规则

- 数据模型
    - metrics name 和 labels
        metrics name：指标名称，如cpu_rate
        label：为指标打上标签，方便查询的时候缩减查询范围
    - samples
        每一次采集收集到的采用数据由两部分组成：
        - TimeStamp
        - Value

    - 指标表示方式
        - OpenTSDB的表示方式
            `<metrics name>{<label name>=<label value>, ...}`，如`api_http_requests_total{method="POST", handler="/messages"}`



- 指标类型
    - Counter：计数器，只增不减
    - Gauge：仪表盘，可增可减
    - Histogram：直方图，最重要的指标，范围时间内对数据进行采样，将其计入可配置的存储桶（提前将可能出现的数值范围进行分段， 均分或者不均分）中，然后按桶来统计出现的频率
    - Summary：摘要，在直方图的基础上，直接统计特定区间的计数

- PromQL
    Prometheus中用于查询数据的语言
    常用：
    ```pql
    histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
    ```
    表示95%的请求在5min内，在不同响应事件区间的处理的数量的变化情况
    
    更多PQL相关信息：[参考](https://prometheus.io/docs/prometheus/latest/querying/basics/)


- 如何在Kubernetes中使用给Prometheu暴露上报指标？
    1. web server应用需要暴露接口`/metrics`来上报指标信息
        ```go
        func register() {
            http.Handle("/metrics", promhttp.HandlerFor(
                registry,
                promhttp.HandlerOpts{
                    EnableOpenMetrics: true,
                }),
            )
        }
        ```
    2. 在Pod Template中需要申明上报指标的端口和地址
        ```yaml
        apiVersion: apps/v1
        kind: Deployment
        metadata:
            name: http-server-demo
        spec:
            replicas: 1
            selector:
                matchLabels:
                app: http-server-demo
            template:
                metadata:
                    labels:
                        app: http-server-demo
                    annotations:
                        prometheus.io/port: server-port
                        prometheus.io/scrape: "true"
                spec:
                    containers:
                        - name: http-server-demo
                        ports:
                            - name: server-port
                            containerPort: 80
        ```

展示监控数据：
- Prometheus Server自带的Dashboard
- Grafana Dashboard
    配置Grafana Dashboard，可以自行定义自己需要的Dashboard，也可以直接使用社区定义好的通用的Dashboard，具体做法是先去社区查找需要的Dashboard的id，然后到grafana dashboard创建界面输入id进行导入
    [Grafana实战](https://zhuanlan.zhihu.com/p/580145725)

配置[告警](https://prometheus.io/docs/alerting/latest/overview/)

- Thanos
当集群规模较大的时候，可以将单个集群的Prometheus汇报的指标汇总到Thanos中，通过sidecar的形式采集集群中Prometheus的数据然后集中汇报
[官网](https://thanos.io/)


# 应用迁移

## Helm
Kubernetes的包管理器，更多[细节](https://helm.sh/)

### 特性
- Helm chart：一个应用实例的必要配置组和，一堆的spec
- 配置信息分为Template和Value，使用这些信息进行渲染，生成最终的对象
- 所有配置可以打包到一个可发布的对象中
- release：一个带有特定配置的Helm chart实例

### 组件
- Helm client
    - 进行本地chart开发与打包压缩文件
    - 管理repository，获取与推送chart
    - 管理release
    - 与liabray交互
        - 发送需要安装的chart
        - 请求升级或者卸载存在的chart
- Helm libaray
    - 与API Server进行交互
        1. 基于chart和configuration创建一个release
        2. 把chart安装到kubernetes集群中，并提供相应的release对象
        3. 升级和卸载
        4. Helm采用Kubernetes存储所有配置信息，无需自己的数据库

### 常用指令
- helm create：创建新的chart
- helm install：安装chart
- helm package：将chart打包为压缩文件
- helm repo 
    - list：查看已经安装的社区repo
    - add: 添加repo
    - update
- helm search `repo/hub`：查找社区存在的chart
- helm pull：拉取chart
- helm upgrade

## Metrics Server

- Metrics Server是Kubernetes监控体系的核心组件之一
- 负载从Kubelet收集资源指标，对这些监控数据进行聚合（依赖kube-aggregator），并在Kubernetes ApiServer中通过Metrics API(`/apis/metrics.k8s.io/`)公开暴露它们。但是metrics-server只会存放最新的指标数据(CPU/Memory)
- 主要是为了驱动集群的弹性伸缩的能力（如HPA）
- 以Aggregated APIServer的形式运行

如何知道特定Group和Version的组合是哪个本地服务还是Aggregate响应的：
```bash
$ kubectl get apiservice
NAME                                   SERVICE                      AVAILABLE   AGE
v1.                                    Local                        True        33d
v1.admissionregistration.k8s.io        Local                        True        33d
v1.apiextensions.k8s.io                Local                        True        33d
v1beta1.metrics.k8s.io                 kube-system/metrics-server   True        3m14s

$ kubectl get svc -n kube-system
NAME             TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                  AGE
kube-dns         ClusterIP   10.96.0.10       <none>        53/UDP,53/TCP,9153/TCP   33d
metrics-server   ClusterIP   10.109.228.178   <none>        443/TCP                  7m30s
```
SERVICE列显示local表示为原生的api server进行响应，上面结果说明group是`metrics.k8s.io`和version是`v1beta1`的资源会交由`kube-system/metrics-server`这个service进行处理

如何查看节点或者Pod的负载情况：
```bash
$ kubectl top nodes
NAME       CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%   
minikube   554m         9%     1350Mi          11% 

$ kubectl top pod loki-grafana-878b5cc5-thfb6
NAME                          CPU(cores)   MEMORY(bytes)   
loki-grafana-878b5cc5-thfb6   2m           95Mi     

$ kubectl top pod loki-grafana-878b5cc5-thfb6
NAME                          CPU(cores)   MEMORY(bytes)   
loki-grafana-878b5cc5-thfb6   2m           95Mi     
```

# 弹性能力

## HPA

`apiVersion: autoscaling/v2`

概念：
- Horizontal Pod AutoScaler，依赖于Metrics Server，根据某些指标在statefulSet，replicaSet，deployments等集合中的Pod数量进行横向动态伸缩
- 多个冲突的HPA规则应用到同一个应用的时候可能会造成无法预期的行为，需要小心维护

查看HPA：
```bash
$ kubectl get hpa
```

配置样例：
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: php-apache
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: php-apache
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 50
  - type: Pods
    pods:
      metric:
        name: packets-per-second
      target:
        type: AverageValue
        averageValue: 1k
  - type: Object
    object:
      metric:
        name: requests-per-second
      describedObject:
        apiVersion: networking.k8s.io/v1
        kind: Ingress
        name: main-route
      target:
        type: Value
        value: 10k
```

算法细节：
 `期望副本数 = ceil(当前副本数 * (当前指标 / 期望指标))`
 
 - 如CPU的当前用量是200m，目标值是100m，那么期望副本数就是`当前副本数 * 200 / 100`，意味着当前副本数需要翻倍；
 - 如CPU的当前用量是50m，目标值是100m，那么期望副本数就是`当前副本数 * 50 / 100`，意味着当前副本数需要减半；
    
容忍值：
    `--horizontal-pod-autoscaler-tolerance`

    如果计算出来的扩缩比例接近1.0，那么根据容忍值的配置（默认0.1），在(1.0 - tolerance, 1.0 + tolerance)的范围内会放弃本次扩缩操作

冷却/延迟支持：
    `--horizontaol-pod-autoscaler-downscale-stablilization`

- 当使用HPA管理副本扩缩容时，可能会因为指标动态变化造成副本数量频繁变化，称为抖动（Thrashing）
- 通过设置缩容冷却时间窗口长度，HPA可以记住过去建议的负载规模，并进队此时间窗口内进行最大规模次数的操作（默认5min）

    
扩容策略：`spec.behavior`

为了保证扩缩容操作可以平滑执行

配置样例：
```yaml
behavior:
  scaleDown:
    policies:
    - type: Pods
      value: 4
      periodSeconds: 60
    - type: Percent
      value: 10
      periodSeconds: 60
```

通过配置behavior来设定：
- 单位时间内，最多只能伸缩特定个数的Pod或者特定比例的Pod数量

存在的问题：
- 基于指标的弹性有滞后性，因为弹性控制器的操作链路很长，从应用负载超出阈值到HPA完成扩容之间的时间差包括：
    - 应用指标数值已经超出阈值
    - HPA定期执行手机指标的滞后
    - HPA 控制Deployment扩容的时间
    - Pod调度的时间
    - 应用启动到达服务就绪的时间

很可能突发流量出现时，还没完成扩容，现有的服务实例就已经被击垮

## VPA

根据容器资源使用率，自行去设置CPU和内存的requests

> 目前VPA还没有生产就绪，慎用

- 组件
    - Vertical Pod AutoScaler
    - VPA Recommender
        - 监视pod，为它们计算新的资源推荐，并将推荐值存放在VPA对象中
        - 使用Metrics-Server的集群中所有Pod的利用率和OOM事件
    - VPA Admission Controller
        - 所有创建Pod的请求都通过该controller
    - VPA Updater
        - 负载Pod实时更新
    - History Storage
        - 存储组件，使用API Server的利用率信息和OOM并将其持久存储

使用建议：
    - VPA更新模式：
        - OFF：不更新，只会在recommender中存放推荐的资源数值
        - AUTO：动态更新Pod的resource request，如果发生变化，会evcit旧的pod重建新的pod
    推荐采取OFF模式，配合Initial模式（如果Pod需要更新或者重建的场合，再去配合recommender中的推荐值更新resource request，而不是让updater直接重建pod）


# Service Mesh

服务和服务之间的网络调用，变为sidecar和sidecar之间的网络调用，将认证，负载均衡，熔断，限流等能力从业务容器转移到sidecar容器中，让业务代码的职责变得更单一

Mesh具有以下能力：
- 适应性
    - 熔断
    - 限流
    - 超时，失败处理
    - 负载均衡
- 服务发现
    - 路由
- 安全和访问控制
    - TLS和证书管理
- 可观测性
    - metrics
    - log
    - tracing
- 部署
- 通讯
    - 多种协议
- 其它
    - 蓝绿发布
    - A/B测试

## Istio

> 入站流量 - 集群内调用 - 出站流量的统一管控的解决方案

功能：
    - 请求路由
    - 服务发现和负载均衡
    - 故障处理
    - 故障注入
    - 规则配置

### 架构

#### Envoy
Istio的数据面
- 线程模式：
    - 单进程多线程
        - 主线程负责协调
            - xDS API(extended discovery service API): 从管理端拉取配置清单，推缓存到thread local的slot中，然后将信息推送到worker线程的tls中
                v2 API使用pb + grpc实现
        - 子线程负责监听，过滤和转发。
            - 当某个连接被监听器接受，那么该链接的整个生命周期会与某个线程绑定
            - 接受到Down Stream请求后，根据main thread推送过来的信息进行比对后，根据配置新起一个请求发到upstream

    - 基于epoll
    - 建议envoy配置的worker数量和所在硬件的线程数一致

#### Istiod
Istio控制平面：
    监听kubernetes的API Server的配置（Istio定义的CRD）和Kubernetes自身的对象（Service，Pod等，在isio中称为status信息），将这些信息组装称为标准的envoy配置，并且使用一个grpc的服务，让sidecar的envoy来获取这些信息，在发生变动的时候向envoy推送新的增量信息

### 流量劫持

- 为用户应用注入sidecar
    - 自动注入：为特定namespace打一个label，在该namespace下面部署Pod的时候，会被Mutating Webhook拦截做变形，将envoy以sidecar的形式注入pod spec
    ```yaml
    apiVersion: v1
    kind: Namespace
    metadata:
        labels:
            kubernetes.io/metadata.name: cloudnative
            istio-injection: enabled
        name: cloudnative
    ```       
    - 手动注入：
        `istioctl kube-inject -f "your yaml file"`

    sidecar的containers如下：
    - istio-init

        一个init-container，在Pod初始化的时候进行iptables等网络信息的配置
    - istio-proxy

        实际运行的是envoy，运行的用户UID是1337的进程

- 集群内数据包流转
    数据包的流转取决于Pod中的iptables的规则
    - 查看Pod信息
        ```bash
        $ kubectl get pods -n cloudnative bff-svc-6fdb6595b8-698vs -owide
        NAME                       READY   STATUS    RESTARTS      AGE   IP            NODE     NOMINATED NODE   READINESS GATES
        bff-svc-6fdb6595b8-698vs   2/2     Running   4 (19m ago)   21d   10.10.65.19   k8s201   <none>           <none>
        ```

        看到Pod在节点k8s201上，ssh到该节点，查看该Pod所对应的在主机上的pid
        ```bash
        $ crictl pods | grep bff-svc-6fdb6595b8-698vs
        cfee095c83c09       19 minutes ago      Ready               bff-svc-6fdb6595b8-698vs                   cloudnative         2                   (default)

        $ crictl inspectp cfee095c83c09 | grep pid
          "pid": "CONTAINER",
        "pid": 3278,
                "pid": 1
                "type": "pid"
                    "getpid",
                    "getppid",
                    "pidfd_open",
                    "pidfd_send_signal",
                    "waitpid",
        ```

        查看该进程的iptables规则
        ```bash
        $ nsenter -t 3278 -n iptables-save
        # Generated by iptables-save v1.8.4 on Mon Aug 28 06:39:51 2023
        *nat
        :PREROUTING ACCEPT [1611:96660]
        :INPUT ACCEPT [1611:96660]
        :OUTPUT ACCEPT [1066:68178]
        :POSTROUTING ACCEPT [1066:68178]
        :ISTIO_INBOUND - [0:0]
        :ISTIO_IN_REDIRECT - [0:0]
        :ISTIO_OUTPUT - [0:0]
        :ISTIO_REDIRECT - [0:0]
        -A PREROUTING -p tcp -j ISTIO_INBOUND
        -A OUTPUT -p tcp -j ISTIO_OUTPUT
        -A ISTIO_INBOUND -p tcp -m tcp --dport 15008 -j RETURN
        -A ISTIO_INBOUND -p tcp -m tcp --dport 15090 -j RETURN
        -A ISTIO_INBOUND -p tcp -m tcp --dport 15021 -j RETURN
        -A ISTIO_INBOUND -p tcp -m tcp --dport 15020 -j RETURN
        -A ISTIO_INBOUND -p tcp -j ISTIO_IN_REDIRECT
        -A ISTIO_IN_REDIRECT -p tcp -j REDIRECT --to-ports 15006
        -A ISTIO_OUTPUT -s 127.0.0.6/32 -o lo -j RETURN
        -A ISTIO_OUTPUT ! -d 127.0.0.1/32 -o lo -p tcp -m tcp ! --dport 15008 -m owner --uid-owner 1337 -j ISTIO_IN_REDIRECT
        -A ISTIO_OUTPUT -o lo -m owner ! --uid-owner 1337 -j RETURN
        -A ISTIO_OUTPUT -m owner --uid-owner 1337 -j RETURN
        -A ISTIO_OUTPUT ! -d 127.0.0.1/32 -o lo -p tcp -m tcp ! --dport 15008 -m owner --gid-owner 1337 -j ISTIO_IN_REDIRECT
        -A ISTIO_OUTPUT -o lo -m owner ! --gid-owner 1337 -j RETURN
        -A ISTIO_OUTPUT -m owner --gid-owner 1337 -j RETURN
        -A ISTIO_OUTPUT -d 127.0.0.1/32 -j RETURN
        -A ISTIO_OUTPUT -j ISTIO_REDIRECT
        -A ISTIO_REDIRECT -p tcp -j REDIRECT --to-ports 15001
        COMMIT
        ```

    - 请求发出
        `-A OUTPUT -p tcp -j ISTIO_OUTPUT`：该规则说明如果一个tcp协议的数据包走到OUTPUT chain中的时候，就将其无条件转到`ISTIO_OUTPUT`中
        在发出请求的时候，进程是业务容器进程，因此`ISTIO_OUTPUT`的所有规则到最后一条之前都不会命中，直到到达兜底规则：`-A ISTIO_OUTPUT -j ISTIO_REDIRECT`，然后将数据包发去`ISTIO_REDIRECT`;

        `-A ISTIO_REDIRECT -p tcp -j REDIRECT --to-ports 15001`将tcp请求无条件转到15001端口（是envoy监听的一个虚地址）
        
        我们查看一下这端口的listener配置
        ```bash
        $ istioctl pc listener -n cloudnative bff-svc-6fdb6595b8-698vs --port 15001 -o json # pc 全称是 proxy-config
        ```
        15001配置如下：
        ```json
        [
            {
                "name": "virtualOutbound",
                "address": {
                    "socketAddress": {
                        "address": "0.0.0.0",
                        "portValue": 15001
                    }
                },
                // omit others
                "useOriginalDst": true,
                "trafficDirection": "OUTBOUND",
                // omit others
            }
        ]
        ```
        这里的信息是说明15001这个listener没有配置路由信息，`useOriginalDst`为true表示还是使用原来的目标地址，然后根据原来的目标地址中的端口，转到相同端口的listener进行处理，假如我们的请求的端口是80，那么去查看80端口的listener的proxy config配置
        ```bash
        ```
        80配置如下：
        ```json
        [
                {
                "name": "0.0.0.0_80",
                "address": {
                    "socketAddress": {
                        "address": "0.0.0.0",
                        "portValue": 80
                    }
                },
                "filterChains": [
                    {
                        "filterChainMatch": {
                            "transportProtocol": "raw_buffer",
                            "applicationProtocols": [
                                "http/1.1",
                                "h2c"
                            ]
                        },
                        "filters": [
                            {
                                "name": "envoy.filters.network.http_connection_manager",
                                "typedConfig": {
                                    "@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
                                    "statPrefix": "outbound_0.0.0.0_80",
                                    "rds": {
                                        "configSource": {
                                            "ads": {},
                                            "initialFetchTimeout": "0s",
                                            "resourceApiVersion": "V3"
                                        },
                                        "routeConfigName": "80"
                                    }
                                }
                            }
        ]
        ```
        `"routeConfigName": "80"`这段表示当进入到该listener之后，就会使用`80`这个路由配置的信息

        查看名字是80的路由信息
        ```bash
        # istioctl pc route -n cloudnative bff-svc-6fdb6595b8-698vs --name=80 -o json       
        ```
        80的路由如下：
        ```json
        [
            {
                "name": "80",
                "virtualHosts": [
                    {
                        "name": "bff-svc.cloudnative.svc.cluster.local:80",
                        "domains": [
                            "bff-svc.cloudnative.svc.cluster.local",
                            "bff-svc",
                            "bff-svc.cloudnative.svc",
                            "bff-svc.cloudnative",
                            "10.10.98.57"
                        ],
                        "routes": [
                            {
                                "name": "default",
                                "match": {
                                    "prefix": "/"
                                },
                                "route": {
                                    "cluster": "outbound|80||bff-svc.cloudnative.svc.cluster.local",
                                    "timeout": "0s",
                                    "retryPolicy": {
                                        "retryOn": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
                                        "numRetries": 2,
                                        "retryHostPredicate": [
                                            {
                                                "name": "envoy.retry_host_predicates.previous_hosts",
                                                "typedConfig": {
                                                    "@type": "type.googleapis.com/envoy.extensions.retry.host.previous_hosts.v3.PreviousHostsPredicate"
                                                }
                                            }
                                        ],
                                        "hostSelectionRetryMaxAttempts": "5",
                                        "retriableStatusCodes": [
                                            503
                                        ]
                                    },
                                    "maxGrpcTimeout": "0s"
                                },
                                "decorator": {
                                    "operation": "bff-svc.cloudnative.svc.cluster.local:80/*"
                                }
                            }
                        ],
                        "includeRequestAttemptCount": true
                    }
                ],
                "validateClusters": false,
                "maxDirectResponseBodySizeBytes": 1048576,
                "ignorePortInHostMatching": true
            }
        ]
        ```
        当域名和domains中的值匹配命中后，就会交由`routes[0].route.cluster`中的`"outbound|80||bff-svc.cloudnative.svc.cluster.local"`cluster进行处理，接下来就要去查看这个cluster对应有哪些IP地址

        ```bash
        $ istioctl pc endpoint -n cloudnative bff-svc-6fdb6595b8-698vs --cluster="outbound|80||bff-svc.cloudnative.svc.cluster.local"
        ENDPOINT           STATUS      OUTLIER CHECK     CLUSTER
        10.10.65.19:80     HEALTHY     OK                outbound|80||bff-svc.cloudnative.svc.cluster.local
        ```
        最后从endpoint列表中根据一定的算法选定一个ip，完成目标选取，负载均衡完成，然后由envoy使用该地址进行请求

        但是向目标地址发送的请求马上又会被规则`-A OUTPUT -p tcp -j ISTIO_OUTPUT`抓住，但是因为此时是envoy发出的请求，也就是PID=1337的进程发出的请求，会命中`-A ISTIO_OUTPUT -m owner --uid-owner 1337 -j RETURN`，接下来的规则就不会再去匹配，从而直接将请求发出Pod

    - 请求接受
        核心规则如下：
        ```
        -A PREROUTING -p tcp -j ISTIO_INBOUND
        -A OUTPUT -p tcp -j ISTIO_OUTPUT
        -A ISTIO_INBOUND -p tcp -m tcp --dport 15008 -j RETURN
        -A ISTIO_INBOUND -p tcp -m tcp --dport 15090 -j RETURN
        -A ISTIO_INBOUND -p tcp -m tcp --dport 15021 -j RETURN
        -A ISTIO_INBOUND -p tcp -m tcp --dport 15020 -j RETURN
        -A ISTIO_INBOUND -p tcp -j ISTIO_IN_REDIRECT
        -A ISTIO_IN_REDIRECT -p tcp -j REDIRECT --to-ports 15006
        ```
        数据包在PREROUTING阶段就被捕获到ISTIO_INBOUND中，然后除了特定的port以为都走到最后一条，直到ISTIO_IN_REDIRECT转到15006这个虚IP的listener，再根据目标端口转到对应的端口的listener，解析cluster之后发现是自己，然后就通过loopback直接localhost发到Pod中的业务容器进程中