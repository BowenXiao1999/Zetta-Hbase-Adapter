# 总结一下目前所做的工作以及之后的安排

代码地址：https://github.com/BowenXiao1999/ZKLib
主要文件：
https://github.com/BowenXiao1999/ZKLib/blob/master/zketcd_lib.go
https://github.com/BowenXiao1999/ZKLib/blob/master/zketcd_lib_test.go


先归纳一下当前的API，用法＋返回值和go zk client保持一致，[参见](https://godoc.org/github.com/samuel/go-zookeeper/zk):　
1. Set
2. GetData
3. Create
4. Delete
5. Exists
6. ExistsW (返回值不包含chan event)
7. GetChildren

Set, GetData, Create, Delete 属于基本功能，不涉及Watches等等，应该是没有太大的问题。

Exists和ExistsW　２个API主要涉及到SetWatches。这一块相关的剥离处理比较麻烦，因为当前要做的Library并没有conn的概念。

目前已经能实现基本的回调处理，可以参考zketcd_lib_test.go里的相关函数(TestWatches)。但实际上实现过程中是有些问题不明白的，暂且不明白原因。可能的技术是会造成一些Go routine的泄露，因为把相关调用参数的context.Cancel()改成了context.TODO()，以及rev版本相关的问题。根据目前的test来看是没有发现的。

总结：
当前Zookeeper依赖去除的大致框架已经有了，主要是需要讨论重点需要哪些API以及他们的robust程度。
虽然说有相关的Unit Test，但是相对来说还是比较贫瘠，不知道在其他条件下是否足够鲁棒。稍微总结一下努力方向

1. 完善UT，从go-zk里收集一些测试用例并且改写
2. 探讨SetWatches的问题，总结一个文档
3. 或者去看Zetta的代码，考虑接入问题以及怎么调用等等

