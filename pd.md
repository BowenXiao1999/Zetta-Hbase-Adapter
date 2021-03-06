# PD 替换 Etcd

我对目前项目架构的最新思考：
首要目的：骗过所有HBase Client, 并用Zetta来作为底层存储

## 模仿HBase Zk
对于HBase Client来说，第一个要骗的点在于向原生zk发起请求的时候。这里我们有zetcd的加持，可以帮助我们解析zk请求，并用etcd存储。表面上看，这部分已经完成了需求。但是，由于Zetta是基于TiKV构建的，那么我们必须引入PD来进行调度。现在就有2个抉择：

1. 保留zetcd的etcd集群，引入PD做调度。
2. 引入PD，同时修改zetcd源码，使用PD提供的存储功能。（目前调研的方向）


第一种方案，zetcd仅仅被用来做一个简单的KV存储(hbase::meta)，因为这个时候需要管理的是TiKV集群，那么PD就能胜任这个角色。会比较浪费，但是也可以通过砍功能来精简。

第二种方案扩充了PD的能力，使得整个项目的架构更简单了。但实际上PD还是通过了Etcd的能力来存储相关信息。

目前在向第二种方案努力，替换掉ztecd的get meta table location 方法，把请求forward到PD，提供一个简单的KV存储功能。同时也发现PD的文档实在是少，只能去源码中找对应API，但是有关简单的KV存储的接口还是很难找。

## 模仿 Region Server
这里需要Zetta节点或者是TiKV节点能提供读写数据的接口了。同时也要商定HBase的格式。不过由于尚未成型，这一部分留在后面考虑。