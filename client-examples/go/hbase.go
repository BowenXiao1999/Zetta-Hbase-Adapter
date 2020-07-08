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
