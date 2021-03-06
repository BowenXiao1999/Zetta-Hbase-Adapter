// test project main.go
// src: https://blog.csdn.net/zhaominpro/article/details/77543543
// https://blog.csdn.net/whereismatrix/article/details/50420099
// 博客园: https://www.cnblogs.com/zhichaoma/p/12640064.html

package main

import (
    "fmt"

    "time"

    "github.com/samuel/go-zookeeper/zk"
)

func main()  {
	var hosts = []string{"localhost:2181"}//server端host
	conn, _, err := zk.Connect(hosts, time.Second*5)
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	var path="/test2"
	var data=[]byte("hello zk")
	var flags=0
	//flags有4种取值：
	//0:永久，除非手动删除
	//zk.FlagEphemeral = 1:短暂，session断开则改节点也被删除
	//zk.FlagSequence  = 2:会自动在节点后面添加序号
	//3:Ephemeral和Sequence，即，短暂且自动添加序号
	var acls=zk.WorldACL(zk.PermAll)//控制访问权限模式
	
	p,err_create:=conn.Create(path,data,int32(flags),acls)
	if err_create != nil {
		fmt.Println(err_create)
		return
	}
	fmt.Println("create:",p)

	// p,err_create:=conn.Delete(path)
	// if err_create != nil {
	// 	fmt.Println(err_create)
	// 	return
	// }
	// fmt.Println("delete:",p)

}