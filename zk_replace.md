# Zookeeper Go 组件化最小实现

## 背景
Zetta项目需要迁移Hbase业务, 而Hbase依赖zk做一些存储配置。
希望使用Go实现一个提供最小功能的zk，并且用Zetta的Cluster支持，能作为一个组件嵌入Hbase的整体迁移。

etcd 实现过：zetcd, 主要提供的能力是保持使用zookeeper的api，但是使用etcd集群。

## HBase Zk 架构介绍
HBase 使用 Zookeeper 做分布式管理服务，来维护集群中所有服务的状态。Zookeeper 维护了哪些 servers 是健康可用的，并且在 server 故障时做出通知。
通常需要 3-5 台服务器来host。

Region Servers 和 在线 HMaster(active HMaster)和 Zookeeper 保持会话(session)。Zookeeper 通过心跳检测来维护所有临时节点(ephemeral nodes)。

每个 Region Server 都会创建一个 ephemeral 节点。HMaster 会监控这些节点来发现可用的 Region Servers，同样它也会监控这些节点是否出现故障。
HMaster 们会竞争创建 ephemeral 节点，而 Zookeeper 决定谁是第一个作为在线 HMaster，保证线上只有一个 HMaster。在线 HMaster(active HMaster) 会给 Zookeeper 发送心跳，不在线的待机 HMaster (inactive HMaster) 会监听 active HMaster 可能出现的故障并随时准备上位。
如果有一个 Region Server 或者 HMaster 出现故障或各种原因导致发送心跳失败，它们与 Zookeeper 的 session 就会过期，这个 ephemeral 节点就会被删除下线，监听者们就会收到这个消息。Active HMaster 监听的是 region servers 下线的消息，然后会恢复故障的 region server 以及它所负责的 region 数据。而 Inactive HMaster 关心的则是 active HMaster 下线的消息，然后竞争上线变成 active HMaster。

![img](https://pic1.zhimg.com/80/v2-9d4069dbe8462a266992dc0a41888540_1440w.jpg)

首次读写操作过程：
1. 客户端从 Zookeeper 那里获取是哪一台 Region Server 负责管理 Meta table。
2. 客户端会查询那台管理 Meta table 的 Region Server，进而获知是哪一台 Region Server 负责管理本次数据请求所需要的 rowkey。客户端会缓存这个信息，以及 Meta table 的位置信息本身。
3. 然后客户端回去访问那台 Region Server，获取数据。

对于以后的的读请求，客户端可以从缓存中直接获取 Meta table 的位置信息(在哪一台 Region Server 上)，以及之前访问过的 rowkey 的位置信息(哪一台 Region Server 上)，除非因为 Region 被迁移了导致缓存失效。这时客户端会重复上面的步骤，重新获取相关位置信息并更新缓存。



Meta Table数据格式：

![img](https://pic1.zhimg.com/80/v2-df270f9aabb60e8f1857ecb73e1402d8_1440w.jpg)

Meta table 是一个特殊的 HBase table，它保存了系统中所有的 region 列表。这张 table 类似一个 b-tree，结构大致如下：

- Key：table, region start key, region id
- Value：region server



总结zk需要的能力：
1. 负责Active Master的选举，保证集群永远有且仅有一个Active Master在线上。
2. 对 Master, Region Server进行上下线管理 (Master独占锁，Region Server上线通知)
3. zk 需要与几乎所有服务器进程保持会话（心跳检测），需要维护所有Server的连接状态，~~负责故障Node的重启~~不负责故障恢复，只是为故障恢复提供信息。
4. zk 需要维护一个 Meta table (特殊的 HBase 表)，包含了集群中所有 regions 的位置信息(寻址入口)。
5. 解析客户端发过来的zk请求，查找到row key的位置信息后返回给客户端（涉及与客户端RPC请求的编解码以及查找过程）。
6. 存储Hbase的schema，包括有哪些table，每个table有哪些column family。

### Zk Region定位
系统如何找到某个row key (或者某个 row key range)所在的region。bigtable 使用三层类似B+树的结构来保存region位置。

第一层是保存zookeeper里面的文件，它持有root region的位置。

第二层root region是.META.表的第一个region其中保存了.META.z表其它region的位置。通过root region，我们就可以访问.META.表的数据。

第三层是.META.，它是一个特殊的表，保存了hbase中所有数据表的region 位置信息。

说明：
1. root region永远不会被split，保证了最需要三次跳转，就能定位到任意region 。
2. .META.表每行保存一个region的位置信息，row key 采用表名+表的最后一样编码而成。
3. 为了加快访问，.META.表的全部region都保存在内存中。假设，.META.表的一行在内存中大约占用1KB。并且每个region限制为128MB。那么上面的三层结构可以保存的region数目为：(128MB/1KB) * (128MB/1KB) = = 2(34)个region
4. client会将查询过的位置信息保存缓存起来，缓存不会主动失效，因此如果client上的缓存全部失效，则需要进行6次网络来回，才能定位到正确的region(其中三次用来发现缓存失效，另外三次用来获取位置信息)。

**Zk 管理 Master, Region Server 的上下线**

### Region Server 上线
master使用zookeeper来跟踪region server状态。当某个region server启动时，会首先在zookeeper上的server目录下建立代表自己的文件，并获得该文件的独占锁。由于master订阅了server 目录上的变更消息，当server目录下的文件出现新增或删除操作时，master可以得到来自zookeeper的实时通知。因此一旦region server上线，master能马上得到消息。

### Region Server 下线
当region server下线时，它和zookeeper的会话断开，zookeeper而自动释放代表这台server的文件上的独占锁。而master不断轮询 server目录下文件的锁状态。如果master发现某个region server丢失了它自己的独占锁，(或者master连续几次和region server通信都无法成功),master就是尝试去获取代表这个region server的读写锁，一旦获取成功，就可以确定：
1 region server和zookeeper之间的网络断开了。
2 region server挂了。
的其中一种情况发生了，无论哪种情况，region server都无法继续为它的region提供服务了，此时master会删除server目录下代表这台region server的文件，并将这台region server的region分配给其它还活着的同志。
如果网络短暂出现问题导致region server丢失了它的锁，那么region server重新连接到zookeeper之后，只要代表它的文件还在，它就会不断尝试获取这个文件上的锁，一旦获取到了，就可以继续提供服务。


### Master 上线
master启动进行以下步骤:
1 从zookeeper上获取唯一一个代码master的锁，用来阻止其它master成为master。
2 扫描zookeeper上的server目录，获得当前可用的region server列表。
3 和2中的每个region server通信，获得当前已分配的region和region server的对应关系。
4 扫描.META.region的集合，计算得到当前还未分配的region，将他们放入待分配region列表。

### Master 下线
由于master只维护表和region的元数据，而不参与表数据IO的过程，master下线仅导致所有元数据的修改被冻结(无法创建删除表，无法修改表的schema，无法进行region的负载均衡，无法处理region上下线，无法进行region的合并，唯一例外的是region的split可以正常进行，因为只有region server参与)，表的数据读写还可以正常进行。因此master下线短时间内对整个hbase集群没有影响。从上线过程可以看到，master保存的 信息全是可以冗余信息（都可以从系统其它地方收集到或者计算出来），因此，一般hbase集群中总是有一个master在提供服务，还有一个以上 的’master’在等待时机抢占它的位置。



## 源码阅读



## 方案
实现前先要弄清楚几个问题：
1. 客户端传过来的请求，Go client/Java client是否有区别？这涉及到编码解码的实现。
2. 数据维护是要用etcd集群还是Zetta有类似的能力？(如果zk不是性能瓶颈，可以考虑采用etcd来维护)
3. 心跳检测，涉及到和Region Server的信息交流，然而我们现在没有实现Region Server。



## 具体个人行动建议
1. 熟悉zk API，事务处理原理
2. 熟悉Hbase 和 zk的交互
3. 熟悉Zetta 接口





