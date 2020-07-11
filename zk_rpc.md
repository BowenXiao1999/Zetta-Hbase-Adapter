# gohbase Zk 交互源码分析

源码：

(需要先自行进hbase shell建表)

```go
package main


import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
	// "github.com/tsuna/gohbase/pb"
)

// https://akbarahmed.com/2012/08/13/hbase-command-line-tutorial/

func init() {

	// 以Stdout为输出，代替默认的stderr
	logrus.SetOutput(os.Stdout)
	// 设置日志等级
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {

	client := gohbase.NewClient("localhost")
	// create a table and column family
	// put a scan
	// Values maps a ColumnFamily -> Qualifiers -> Values.
	values := map[string]map[string][]byte{"cf": map[string][]byte{"a": []byte("Hello Word")}}
	putRequest, _ := hrpc.NewPutStr(context.Background(), "table", "15", values)
	client.Put(putRequest)

	// Perform a get for the cell with key "15", column family "cf" and qualifier "a"
	family := map[string][]string{"cf": []string{"a"}}
	getRequest, _ := hrpc.NewGetStr(context.Background(), "table", "15",
		hrpc.Families(family))
	getRsp, _ := client.Get(getRequest)
	fmt.Println(getRsp)
}
```





日志：

![image-20200710203244213](/Users/bytedance/Library/Application Support/typora-user-images/image-20200710203244213.png)

从这个源码中可以分析hbase的读写流程：

1. 第一次look up region的时候，region 被封装成要找table名是myTable。
2. 当发现这个请求需要reestablishing region时（意味着要找zk查hbase:meta），第二个请求返回了hbase:meta的Region。这时候会调用zklookup。
3. 调用zklookup会尝试和zk建立连接，2181正好是HBase组件zk的默认端口号。
4. 连接上zk后查询hbase:meta，这时候会返回一个hbase:meat的region client。这个region client代表已经和负责本次I/O的Server连接上了。因为16020正好是负责客户端接入的HBase Region Server端口号。这个region client负责所有client的I/O
5. 用这个region client去发起新的region look up，这时候要找的region info又换成了myTable。这时候会显示added new region，有可能代表新分配了一块region？
6. 不要忘了这是一个Put请求，但是由于现在region client已经在cache里了（table对应的region也已经建立），那么这次Put就会轻松成功(Flushing multiRequest)。
7. 还有一个Flushing MultiRequest代表Get，直接从缓存里拿了结果。





结合一些源码分析一下：

1. 

Put -> mutate -> SendRPC

![image-20200710182308689](/Users/bytedance/Library/Application Support/typora-user-images/image-20200710182308689.png)

RPC的过程：先拿Region，然后给Region发RPC。





2. 

![image-20200710182355563](/Users/bytedance/Library/Application Support/typora-user-images/image-20200710182355563.png)

拿Region先从缓存查



3. 

终于可以看到与Zk的交互了，可以看到zk负责查谁是Master，hbase:meta在哪里。

![image-20200710182655060](/Users/bytedance/Library/Application Support/typora-user-images/image-20200710182655060.png)



4. 深入zkLookup的源码看一下

   





老师，我总结了一下我的疑问，您有空的时候看一下就行

根据昨天发过来的zkcli.go，是对go zk client的封装。那么我就先简单假设当前项目中有用到 Java Zk client和 Go Zk Client。但是这其实都不是问题，我深入研究了zetcd，它适配很多Zk API，比如Java和Go，这些Client对zetcd的存在是不感知的，它们只知道这是一个能提供Zk能力的服务。

**所以我的理解是我们的业务代码不需要任何改动，只要本地起zetcd来替换原有的Zk节点，就可以切换到etcd。** 



但这并不是完全的解决方案，有可能

1. 某些版本API不支持（未调查，不过几率低）
2. 公司运维体系原因，不希望用etcd集群来支撑这个服务（几率高）
3. 觉得zk的服务过多，希望一个精简版



那么基于以上的分析，对于HBase适配Zetta这个项目，就比较容易得出接下来的方向：

1. 先用zetcd代理HBase中的Zk请求
2. 后期尝试改进zetcd，用pd或者其他存储服务来代替。
3. 思考如何砍掉一些实现 （但是其实用到的功能真的少吗？我认为在HBase里Zk几乎所有的特性都用到了，也就是说我们后期要支持一个fully HBase，Zk的功能还是尽可能完全一点比较好）



那么基于以上方向，来看一下我的疑惑或者理解，

> 就能处理掉 访问 zk 的请求就好了

zetcd目前可以做到支持zk Java/go client api的请求，。



> 搞清楚 hbase client 的访问流程，接下来就尝试用 zetcd 构造一个 小 zk 就行，实现基本 crud 接口就好

没懂这句话，既然我们是用zetcd已经起了server来代替Zk请求，为什么还需要这个zetcd的api来实现来绕一圈呢？我看了zetcd的api，和zookeeper非常像，可以说就是按一种数据模型设计的。2种方式最终都是用etcd维护数据，所以我觉得区别不大。





> 1. Zookeeper 节点存储格式协议的节点，这样原生 hbase client 能够
>    获得正确的信息
>
> 2. 实现类似 zetcd 那样的系统，能够替代 zookeeper 向 hbase client 提供访问
> 3. 剥离相关 的 Zookeeper 通信模块，用 pd 存储或者其他方式来代替 zookeeper 的后端存储

前2点都可以用zetcd代理zk请求来解决。第三点需要对zetcd进行改动，思考如何把zk的数据格式放到pd上存。



所以就是我认为的解决方案是

1. 利用开源的zetcd，适配zk接口，给予etcd支持
2. 慢慢替换相关存储方式，需要重新设计zk节点信息的存储（比如在pd怎么存）。