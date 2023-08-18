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
为每个微服务设置一个总负责人，需要为应用的全生命周期负责。
一个应用ready，需要包括以下内容：
- function ready，业务功能完善
- prodution ready
    - 通过负载与压力测试
    - 完成用户手册
    - 完成管理手册，可以按照管理手册部署与升级服务


