# 总结一下目前所做的工作以及之后的安排

代码地址：https://github.com/BowenXiao1999/ZKLib
主要文件：
https://github.com/BowenXiao1999/ZKLib/blob/master/zketcd_lib.go
https://github.com/BowenXiao1999/ZKLib/blob/master/zketcd_lib_test.go


首先归纳一下当前的API，【用法 + 返回值 + 函数调用参数】和go zk client保持一致，[参见](https://godoc.org/github.com/samuel/go-zookeeper/zk):　
1. Set
2. GetData
3. Create
4. Delete
5. Exists
6. ExistsW (返回值不包含chan event)
7. GetChildren

Set, GetData, Create, Delete 属于基本功能，不涉及Watches等等，没有太大的问题。

Exists和ExistsW　２个API主要涉及到SetWatches。这一块相关的剥离处理比较麻烦，因为当前要做的Library并没有conn的概念。我们需要把可能引起Bug的代码删除修改。

问题主要集中在Watches的一些细节。

## Watches的小问题
目前已经能实现Watches相关的回调处理，可以参考zketcd_lib_test.go里的相关函数(TestWatches)。但实际上实现过程中是有些问题不明白的，暂且不明白原因。

当前版本可能会造成一些Go routine的泄露，因为把相关调用参数的context.Cancel()改成了context.TODO()。虽然根据目前的test来看是没有发现的，但是不排除在高调用频率的情况下会出问题。

同时对于相关数据的rev版本，这个是etcd内部维护的实现MVCC的版本号。为了达到完全模拟rev的效果，我们需要在library内部实现完全模拟rev。目前这一块的支持还不够。

以及对于我测试Watches相关函数的结果（相关UT），我发现调用出现了比较大的延迟现象，以至于回调函数偶尔来不及被执行。因此我在代码中加入了一些Sleep来等待足够长的时间来确保能看到回调函数的输出。可能原因暂时不明，但是我认为原因有可能和Golang的Goroutine执行机制有关，回调函数可能不和主routine在一个routine执行。

## 总结

当前Zookeeper依赖去除的大致框架已经有了，主要是需要讨论重点需要哪些API以及需要支持它们到哪种程度。

虽然说有相关的Unit Test，但是相对来说还是比较贫瘠，不知道在其他条件下是否足够鲁棒。稍微总结一下努力方向：

1. 完善UT，从go-zk里收集一些测试用例并且改写
2. 探讨SetWatches的问题，总结一个文档
3. 或者去看Zetta的代码，考虑接入问题以及怎么调用等等

