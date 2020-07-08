# Zookeeper Go 组件化最小实现

## 背景
Zetta项目需要迁移Hbase业务, 而Hbase依赖zk做一些存储配置。
希望使用Go实现一个提供最小功能的zk，并且用Zetta的Cluster支持，能作为一个组件嵌入Hbase的整体迁移。

etcd 实现过：zetcd, 主要提供的能力是保持使用zookeeper的api，但是使用etcd集群。

## 方案
整体来说，实现分为2个部分

1. 熟悉Hbase在运行时和zk需要的交互过程，了解相关接口


2. Back with Zetta Cluster，使用Zetta集群提供的接口，解析zk写入信息


## 具体个人行动建议
1. 熟悉zk API，事务处理原理
2. 熟悉Hbase 和 zk的交互
3. 熟悉Zetta 接口





