package main

import (
	// "bytes"
	// "encoding/binary"
	"fmt"
	// "log"
	// "math/rand"
	// "path"
	// "path/filepath"
	// "strings"
	// "time"

	// "github.com/davecgh/go-spew/spew"
	// "github.com/golang/protobuf/proto"
	// "github.com/samuel/go-zookeeper/zk"

)




func main()  {
	// new a client
	zkCli := NewZookeeperClient([]string{"localhost:2181"}, "cluster", "root");

	// simple read write case
	zkCli.SetMetaRegionServer("localhost", "2181", 1);


	meta, err:= zkCli.getMetaRegionServer();
	if err == nil {
		fmt.Println(meta)
	}
	
}

