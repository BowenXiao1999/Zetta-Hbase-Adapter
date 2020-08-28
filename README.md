# Zetta-Hbase-Adapter

Record for Zetta Hbase adaption Project.



## Goals

### Week 1

- [x] Install Standalone Hbase
  
- [x] Install Zookeeper

- [x] Use Go client to manipulate Hbase and Zookeeper (client-exmpales)

- [x] Use Java Client to manipulate Hbase and Zookeeper (client-exmpales)


- [x] Investigate the details of zk in Hbase and Come up a [document](./zk_replace.md) for Implementation. 

### Week 2

- [x] Read through write/read source code of hbase client [document](./zk_rpc.md)

- [x] Deploy a zetcd and test it with go-hbase [document](./hbase_adapt.md)


- [x] Try to build zetcd after change source code (fail for a long time) [document](./rebuild.md)

- [x] Test rebuild zetcd with CRUD Request [document](./zk_create.md)

- [x] Analyze the source code of zetcd in depth and evaluate the performance [document](./zetcd_performance.md)
 


- [x] Eliminate the dependency of etcd for GetData(/hbase/meta-region-server) [document](./zk_replace2.md)

### Week3 
- [x] Relace back cluster etcd by pd [document](./pd_replace.md)

- [x] Finish the basic implementation of ZKLib and Test it 

### Week4
- [x] Support basic API of (SetData/Get/Create/Delete)

- [x] Implement basic ExistW to setWatches


### Week5-6
- [x] Enhance the robust of API (SetData/Get/Create/Delete/SetWatches)

- [x] Enrich the Unit Test and refactor it

### Week7
- [ ] Summary my work into a [document]()




## Useful Links
### HBase Client RPC 相关资料

```
https://blog.csdn.net/iteye_14085/article/details/82479582?utm_medium=distribute.pc_relevant_t0.none-task-blog-BlogCommendFromMachineLearnPai2-1.nonecase&depth_1-utm_source=distribute.pc_relevant_t0.none-task-blog-BlogCommendFromMachineLearnPai2-1.nonecase

https://blog.csdn.net/iteye_14085/article/details/82479437

https://www.jianshu.com/p/6c5ea570e70a

https://blog.csdn.net/vovo2008/article/details/84354230?ops_request_misc=%257B%2522request%255Fid%2522%253A%2522159153227419725222441664%2522%252C%2522scm%2522%253A%252220140713.130102334.pc%255Fall.%2522%257D&request_id=159153227419725222441664&biz_id=0&utm_medium=distribute.pc_search_result.none-task-blog-2~all~first_rank_ecpm_v1~pc_rank_v3-4-84354230.first_rank_ecpm_v1_pc_rank_v3&utm_term=hbase+%E9%80%9A%E4%BF%A1%E5%8D%8F%E8%AE%AE

https://blog.csdn.net/chixian2520/article/details/100727890

https://www.jianshu.com/p/01ffc4178c43

http://wenda.chinahadoop.cn/article/23

https://wenda.chinahadoop.cn/article/22
```

### Zookeeper 协议相关项目
```
https://github.com/etcd-io/zetcd
```

### 适配对标的 HBase 版本 
```
https://archive.cloudera.com/cdh5/ubuntu/xenial/amd64/cdh/pool/contrib/h/hbase/hbase_1.2.0+cdh5.14.0+440.orig.tar.gz
```