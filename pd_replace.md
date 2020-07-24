# 替换Backend 为PD集群

在[AskTUG](https://asktug.com/t/topic/36076)上得到关于etcd client 和PD比较肯定的消息后，了解到了PD集群其实支持etcd的所有API，这也大大减少了接入难度。以下为测试细节

## 启动PD
Client URL 在2379

```
[2020/07/24 15:40:28.010 +08:00] [INFO] [util.go:49] ["Welcome to Placement Driver (PD)"]
[2020/07/24 15:40:28.010 +08:00] [INFO] [util.go:50] [PD] [release-version=v4.0.0-rc.2-68-g1ad59bcb]
[2020/07/24 15:40:28.010 +08:00] [INFO] [util.go:51] [PD] [edition=Community]
[2020/07/24 15:40:28.010 +08:00] [INFO] [util.go:52] [PD] [git-hash=1ad59bcbf36d87082c79a1fffa3b0895234ac862]
[2020/07/24 15:40:28.010 +08:00] [INFO] [util.go:53] [PD] [git-branch=master]
[2020/07/24 15:40:28.010 +08:00] [INFO] [util.go:54] [PD] [utc-build-time="2020-07-18 05:32:27"]
[2020/07/24 15:40:28.010 +08:00] [INFO] [metricutil.go:81] ["disable Prometheus push client"]
[2020/07/24 15:40:28.010 +08:00] [INFO] [server.go:210] ["PD Config"] [config="{\"client-urls\":\"http://127.0.0.1:2379\",\"peer-urls\":\"http://127.0.0.1:2380\",\"advertise-client-urls\":\"http://127.0.0.1:2379\",\"advertise-peer-urls\":\"http://127.0.0.1:2380\",\"name\":\"pd-C02YV0JFLVCL\",\"data-dir\":\"/Users/bytedance/code/pd/default.pd-C02YV0JFLVCL\",\"force-new-cluster\":false,\"enable-grpc-gateway\":true,\"initial-cluster\":\"pd-C02YV0JFLVCL=http://127.0.0.1:2380\",\"initial-cluster-state\":\"new\",\"join\":\"\",\"lease\":3,\"log\":{\"level\":\"\",\"format\":\"text\",\"disable-timestamp\":false,\"file\":{\"filename\":\"\",\"max-size\":0,\"max-days\":0,\"max-backups\":0},\"development\":false,\"disable-caller\":false,\"disable-stacktrace\":false,\"disable-error-verbose\":true,\"sampling\":null},\"tso-save-interval\":\"3s\",\"metric\":{\"job\":\"pd-C02YV0JFLVCL\",\"address\":\"\",\"interval\":\"15s\"},\"schedule\":{\"max-snapshot-count\":3,\"max-pending-peer-count\":16,\"max-merge-region-size\":20,\"max-merge-region-keys\":200000,\"split-merge-interval\":\"1h0m0s\",\"enable-one-way-merge\":\"false\",\"enable-cross-table-merge\":\"false\",\"patrol-region-interval\":\"100ms\",\"max-store-down-time\":\"30m0s\",\"leader-schedule-limit\":4,\"leader-schedule-policy\":\"count\",\"region-schedule-limit\":2048,\"replica-schedule-limit\":64,\"merge-schedule-limit\":8,\"hot-region-schedule-limit\":4,\"hot-region-cache-hits-threshold\":3,\"store-limit\":null,\"tolerant-size-ratio\":0,\"low-space-ratio\":0.8,\"high-space-ratio\":0.7,\"scheduler-max-waiting-operator\":5,\"enable-remove-down-replica\":\"true\",\"enable-replace-offline-replica\":\"true\",\"enable-make-up-replica\":\"true\",\"enable-remove-extra-replica\":\"true\",\"enable-location-replacement\":\"true\",\"enable-debug-metrics\":\"false\",\"schedulers-v2\":[{\"type\":\"balance-region\",\"args\":null,\"disable\":false,\"args-payload\":\"\"},{\"type\":\"balance-leader\",\"args\":null,\"disable\":false,\"args-payload\":\"\"},{\"type\":\"hot-region\",\"args\":null,\"disable\":false,\"args-payload\":\"\"},{\"type\":\"label\",\"args\":null,\"disable\":false,\"args-payload\":\"\"}],\"schedulers-payload\":null,\"store-limit-mode\":\"manual\"},\"replication\":{\"max-replicas\":3,\"location-labels\":\"\",\"strictly-match-label\":\"false\",\"enable-placement-rules\":\"false\"},\"pd-server\":{\"use-region-storage\":\"true\",\"max-gap-reset-ts\":\"24h0m0s\",\"key-type\":\"table\",\"runtime-services\":\"\",\"metric-storage\":\"\",\"dashboard-address\":\"auto\"},\"cluster-version\":\"0.0.0\",\"quota-backend-bytes\":\"8GiB\",\"auto-compaction-mode\":\"periodic\",\"auto-compaction-retention-v2\":\"1h\",\"TickInterval\":\"500ms\",\"ElectionInterval\":\"3s\",\"PreVote\":true,\"security\":{\"cacert-path\":\"\",\"cert-path\":\"\",\"key-path\":\"\",\"cert-allowed-cn\":null},\"label-property\":null,\"WarningMsgs\":null,\"DisableStrictReconfigCheck\":false,\"HeartbeatStreamBindInterval\":\"1m0s\",\"LeaderPriorityCheckInterval\":\"1m0s\",\"dashboard\":{\"tidb_cacert_path\":\"\",\"tidb_cert_path\":\"\",\"tidb_key_path\":\"\",\"public_path_prefix\":\"\",\"internal_proxy\":false,\"disable_telemetry\":false},\"replication-mode\":{\"replication-mode\":\"majority\",\"dr-auto-sync\":{\"label-key\":\"\",\"primary\":\"\",\"dr\":\"\",\"primary-replicas\":0,\"dr-replicas\":0,\"wait-store-timeout\":\"1m0s\",\"wait-sync-timeout\":\"1m0s\"}}}"]
```



## 启动zetcd

forward 到2379端口

```
C02YV0JFLVCL:zetcd bytedance$ ./bin/zetcd --zkaddr 0.0.0.0:2181 --endpoints localhost:2379
This my own zetcd: Running zetcd proxy
Version: v0.0.5-ac03d75c618e01b59532ab7dbd3a755287591af4-dirty
SHA: ac03d75c618e01b59532ab7dbd3a755287591af4
log
I'm In server.go
```





## 启动Hbase

```
(base) C02YV0JFLVCL:bin bytedance$ ./start-hbase.sh 
Java HotSpot(TM) 64-Bit Server VM warning: Ignoring option UseConcMarkSweepGC; support was removed in 14.0
running master, logging to /usr/local/var/log/hbase/hbase-bytedance-master-C02YV0JFLVCL.out
Java HotSpot(TM) 64-Bit Server VM warning: Ignoring option UseConcMarkSweepGC; support was removed in 14.0
WARNING: An illegal reflective access operation has occurred
WARNING: Illegal reflective access by org.apache.hadoop.hbase.util.UnsafeAvailChecker (file:/usr/local/Cellar/hbase/2.2.3/libexec/lib/hbase-common-2.2.3.jar) to method java.nio.Bits.unaligned()
WARNING: Please consider reporting this to the maintainers of org.apache.hadoop.hbase.util.UnsafeAvailChecker
WARNING: Use --illegal-access=warn to enable warnings of further illegal reflective access operations
WARNING: All illegal access operations will be denied in a future release
: running regionserver, logging to /usr/local/var/log/hbase/hbase-bytedance-regionserver-C02YV0JFLVCL.out
: Java HotSpot(TM) 64-Bit Server VM warning: Ignoring option UseConcMarkSweepGC; support was removed in 14.0
: WARNING: An illegal reflective access operation has occurred
: WARNING: Illegal reflective access by org.apache.hadoop.hbase.util.UnsafeAvailChecker (file:/usr/local/Cellar/hbase/2.2.3/libexec/lib/hbase-common-2.2.3.jar) to method java.nio.Bits.unaligned()
: WARNING: Please consider reporting this to the maintainers of org.apache.hadoop.hbase.util.UnsafeAvailChecker
: WARNING: Use --illegal-access=warn to enable warnings of further illegal reflective access operations
: WARNING: All illegal access operations will be denied in a future release
```





## 跑例子

```
(base) C02YV0JFLVCL:examples bytedance$ go run hbase.go 
DEBU[0000] Creating new client.                          Host=localhost
DEBU[0000] looking up region                             key="\"15\"" table="\"myTable2\""
DEBU[0000] reestablishing region                         region="RegionInfo{Name: \"hbase:meta,,1\", ID: 0, Namespace: \"hbase\", Table: \"meta\", StartKey: \"\", StopKey: \"\"}"
DEBU[0000] looking up region server of hbase:meta        resource=/meta-region-server
DEBU[0000] Connected to [::1]:2181                      
DEBU[0000] authenticated: id=7668631258093993422, timeout=30000 
DEBU[0000] re-submitting `0` credentials after reconnect 
DEBU[0000] recv loop terminated: err=EOF                
DEBU[0000] send loop terminated: err=<nil>              
DEBU[0000] looked up a region                            addr="10.91.44.122:16020" key="\"\"" region="RegionInfo{Name: \"hbase:meta,,1\", ID: 0, Namespace: \"hbase\", Table: \"meta\", StartKey: \"\", StopKey: \"\"}" table="\"hbase:meta\""
INFO[0000] added new region client                       client="RegionClient{Addr: 10.91.44.122:16020}"
DEBU[0000] looked up a region                            addr="10.91.44.122:16020" key="\"15\"" region="RegionInfo{Name: \"myTable2,,1595069929674.5dbaa81e5f79b158b6f25bc644ceab8c.\", ID: 1595069929674, Namespace: \"\", Table: \"myTable2\", StartKey: \"\", StopKey: \"\"}" table="\"myTable2\""
INFO[0000] added new region                              overlaps="[]" region="RegionInfo{Name: \"myTable2,,1595069929674.5dbaa81e5f79b158b6f25bc644ceab8c.\", ID: 1595069929674, Namespace: \"\", Table: \"myTable2\", StartKey: \"\", StopKey: \"\"}" replaced=true
DEBU[0000] region client is already in client's cache    client="RegionClient{Addr: 10.91.44.122:16020}"
DEBU[0000] flushing MultiRequest                         addr="10.91.44.122:16020" len=1
DEBU[0000] flushing MultiRequest                         addr="10.91.44.122:16020" len=1
cells:[row:"15"  family:"cf1"  qualifier:"a"  timestamp:1595576773440  cell_type:PUT  value:"Hello Word"] stale:false partial:false exists:<nil> 
```

正常。

关闭PD后重新测试，失败。

```
(base) C02YV0JFLVCL:examples bytedance$ go run hbase.go 
DEBU[0000] Creating new client.                          Host=localhost
DEBU[0000] looking up region                             key="\"15\"" table="\"myTable2\""
DEBU[0000] reestablishing region                         region="RegionInfo{Name: \"hbase:meta,,1\", ID: 0, Namespace: \"hbase\", Table: \"meta\", StartKey: \"\", StopKey: \"\"}"
DEBU[0000] looking up region server of hbase:meta        resource=/meta-region-server
DEBU[0000] Connected to [::1]:2181                      
DEBU[0000] authenticated: id=7668631258093993422, timeout=30000 
DEBU[0000] re-submitting `0` credentials after reconnect 
DEBU[0000] recv loop terminated: err=EOF                
DEBU[0000] send loop terminated: err=<nil>              
DEBU[0000] looked up a region                            addr="10.91.44.122:16020" key="\"\"" region="RegionInfo{Name: \"hbase:meta,,1\", ID: 0, Namespace: \"hbase\", Table: \"meta\", StartKey: \"\", StopKey: \"\"}" table="\"hbase:meta\""
INFO[0000] added new region client                       client="RegionClient{Addr: 10.91.44.122:16020}"
DEBU[0000] looked up a region                            addr="10.91.44.122:16020" key="\"15\"" region="RegionInfo{Name: \"myTable2,,1595069929674.5dbaa81e5f79b158b6f25bc644ceab8c.\", ID: 1595069929674, Namespace: \"\", Table: \"myTable2\", StartKey: \"\", StopKey: \"\"}" table="\"myTable2\""
INFO[0000] added new region                              overlaps="[]" region="RegionInfo{Name: \"myTable2,,1595069929674.5dbaa81e5f79b158b6f25bc644ceab8c.\", ID: 1595069929674, Namespace: \"\", Table: \"myTable2\", StartKey: \"\", StopKey: \"\"}" replaced=true
DEBU[0000] region client is already in client's cache    client="RegionClient{Addr: 10.91.44.122:16020}"
DEBU[0000] flushing MultiRequest                         addr="10.91.44.122:16020" len=1
DEBU[0000] flushing MultiRequest                         addr="10.91.44.122:16020" len=1
cells:[row:"15"  family:"cf1"  qualifier:"a"  timestamp:1595576773440  cell_type:PUT  value:"Hello Word"] stale:false partial:false exists:<nil> 
(base) C02YV0JFLVCL:examples bytedance$ go run hbase.go 
DEBU[0000] Creating new client.                          Host=localhost
DEBU[0000] looking up region                             key="\"15\"" table="\"myTable2\""
DEBU[0000] reestablishing region                         region="RegionInfo{Name: \"hbase:meta,,1\", ID: 0, Namespace: \"hbase\", Table: \"meta\", StartKey: \"\", StopKey: \"\"}"
DEBU[0000] looking up region server of hbase:meta        resource=/meta-region-server
DEBU[0000] Connected to [::1]:2181                      
ERRO[0030] failed looking up region                      backoff=16ms err="context deadline exceeded" key="\"\"" table="\"hbase:meta\""
ERRO[0030] failed looking up region                      backoff=16ms err="context deadline exceeded" key="\"15\"" table="\"myTable2\""
DEBU[0030] looking up region server of hbase:meta        resource=/meta-region-server
DEBU[0030] looking up region                             key="\"15\"" table="\"myTable2\""
DEBU[0030] Connected to [::1]:2181                      
ERRO[0060] failed looking up region                      backoff=32ms err="context deadline exceeded" key="\"\"" table="\"hbase:meta\""
ERRO[0060] failed looking up region                      backoff=32ms err="context deadline exceeded" key="\"15\"" table="\"myTable2\""
DEBU[0060] looking up region                             key="\"15\"" table="\"myTable2\""
DEBU[0060] looking up region server of hbase:meta        resource=/meta-region-server
DEBU[0060] Connected to [::1]:2181   
```



## 总结

PD对Etcd的请求做了处理，兼容所有API。目前HBase替换已完成。