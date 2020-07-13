package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/samuel/go-zookeeper/zk"

	"hregion/hserver/pb"
)

var (
	// base znode for this cluster
	BaseZNode string
	// znode containing location of server hosting meta region
	MetaServerZNode string = "meta-region-server"
	// znode containing ephemeral nodes of the regionservers
	RSZNode string = "rs"
	// znode containing ephemeral nodes of the draining regionservers
	DrainingZNode = "draining"
	// znode of currently active master
	MasterAddressZNode string = "master"
	// znode of this master in backup master directory, if not the active master
	BackupMasterAddressesZNode = "backup-master"
	// znode containing the current cluster state
	ClusterStateZNode = "running"
	// znode used for region transitioning and assignment
	AssignmentZNode string = ""
	// znode used for table disabling/enabling
	TableZNode string = "table"
	// znode containing the unique cluster ID
	ClusterIdZNode string = "hbaseid"
	// znode used for log splitting work assignment
	SplitLogZNode = "splitWAL"
	// znode containing the state of the load balancer
	BalancerZNode = "balancer"
	// znode containing the lock for the tables
	TableLockZNode string = "table-lock"
	// znode containing the state of recovering regions
	RecoveringRegionsZNode string = "recovering-regions"
	// znode containing namespace descriptors
	NamespaceZNode string = "namespace"

	defaultVersion int32 = -1
)

func init() {
	zk.DefaultLogger = defaultLogger{}
}

type ZookeeperClient struct {
	hosts   []string
	cluster string
	root    string
	conn    *zk.Conn
	salter  *rand.Rand
}

type defaultLogger struct{}

func (defaultLogger) Printf(format string, a ...interface{}) {
	log.Output(3, fmt.Sprintf(format, a...))
}

func NewZookeeperClient(hosts []string, cluster, root string) (*ZookeeperClient, error) {
	zkconn, _, err := zk.Connect(hosts, 10*time.Second)
	if err != nil {
		return nil, err
	}
	salter := rand.New(rand.NewSource(time.Now().Unix()))
	client := &ZookeeperClient{
		hosts:   hosts,
		cluster: cluster,
		root:    root,
		conn:    zkconn,
		salter:  salter,
	}

	return client, nil
}

func (zk *ZookeeperClient) Stop() {
	zk.conn.Close()
}

func (zk *ZookeeperClient) ListRegionServers() ([]string, error) {
	path := filepath.Join("/", zk.root, "rs")
	regionServers, _, err := zk.conn.Children(path)
	return regionServers, err
}

func (zk *ZookeeperClient) GetRegionServerData(rs string) (*pb.RegionServerInfo, error) {
	path := filepath.Join("/", zk.root, RSZNode, rs)
	buf, _, err := zk.conn.Get(path)
	if err != nil {
		return nil, err
	}
	body, err := GetPayload(path, buf)
	if err != nil {
		return nil, err
	}
	regionSvr := &pb.RegionServerInfo{}
	err = proto.Unmarshal(body, regionSvr)
	if err != nil {
		return nil, err
	}
	return regionSvr, nil
}

func (zk *ZookeeperClient) GetMasterData() (*pb.Master, error) {
	path := filepath.Join("/", zk.root, MasterAddressZNode)
	buf, _, err := zk.conn.Get(path)
	if err != nil {
		return nil, err
	}

	body, err := GetPayload(path, buf)
	if err != nil {
		return nil, err
	}
	master := &pb.Master{}
	err = proto.Unmarshal(body, master)
	if err != nil {
		return nil, err
	}
	return master, nil
}

func (zk *ZookeeperClient) SetMaster(host string, port uint32, startCode uint64) error {
	masterProto := &pb.Master{
		Master: &pb.ServerName{
			HostName:  proto.String(host),
			Port:      proto.Uint32(port),
			StartCode: proto.Uint64(startCode),
		},
		RpcVersion: proto.Uint32(0),
		InfoPort:   proto.Uint32(60010),
	}
	data, err := proto.Marshal(masterProto)
	if err != nil {
		return err
	}
	if err := zk.SetNode(MasterAddressZNode, data); err != nil {
		return err
	}
	return nil
}

func (zk *ZookeeperClient) GetMetaRegionServer() (*pb.MetaRegionServer, error) {
	path := filepath.Join("/", zk.root, MetaServerZNode)
	buf, _, err := zk.conn.Get(path)
	if err != nil {
		return nil, err
	}

	body, err := GetPayload(path, buf)
	if err != nil {
		return nil, err
	}
	metaServer := &pb.MetaRegionServer{}
	err = proto.Unmarshal(body, metaServer)
	if err != nil {
		return nil, err
	}
	return metaServer, nil
}

func (zk *ZookeeperClient) SetMetaRegionServer(host string, port uint32, startCode uint64) error {
	metaSvrProto := &pb.MetaRegionServer{
		Server: &pb.ServerName{
			HostName:  proto.String(host),
			Port:      proto.Uint32(port),
			StartCode: proto.Uint64(startCode),
		},
		RpcVersion: proto.Uint32(0),
		State:      pb.RegionState_OPEN.Enum(),
	}
	data, err := proto.Marshal(metaSvrProto)
	if err != nil {
		return err
	}
	if err := zk.SetNode(MetaServerZNode, data); err != nil {
		return err
	}
	return nil
}

func (zk *ZookeeperClient) SetNode(resource string, pureData []byte) error {
	path := filepath.Join("/", zk.root, resource)
	newData := zk.AppendMetaData(PrependPBMagic(pureData))
	if err := zk.CreateOrUpdate(path, newData, 0); err != nil {
		return err
	}
	return nil
}

func (zk *ZookeeperClient) setData(path string, buffer []byte) error {
	newData := zk.AppendMetaData(buffer)
	_, err := zk.conn.Set(path, newData, -1)
	if err != nil {
		return err
	}
	return nil
}

func (zk *ZookeeperClient) GetNode(resource string) ([]byte, error) {
	path := filepath.Join("/", zk.root, resource)
	buf, _, err := zk.conn.Get(path)
	if err != nil {
		return nil, err
	}
	spew.Dump(buf)
	return buf, nil
}

func (zk *ZookeeperClient) AppendMetaData(data []byte) []byte {
	if data == nil || len(data) == 0 {
		return data
	}
	salt := zk.salter.Uint32()
	idLength := len(zk.cluster) + 4
	bytesBuffer := bytes.NewBuffer([]byte{})
	bytesBuffer.WriteByte(0xFF)
	if err := binary.Write(bytesBuffer, binary.BigEndian, int32(idLength)); err != nil {
		log.Println(err)
		return nil
	}
	bytesBuffer.Write([]byte(zk.cluster))
	if err := binary.Write(bytesBuffer, binary.BigEndian, salt); err != nil {
		log.Println(err)
		return nil
	}
	bytesBuffer.Write(data)
	return bytesBuffer.Bytes()
}

func PrependPBMagic(data []byte) []byte {
	PBMagic := []byte("PBUF")
	return append(PBMagic, data...)
}

func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func GetPayload(resource string, buffer []byte) ([]byte, error) {
	var buf []byte = buffer
	spew.Dump(buf)
	if len(buf) == 0 {
		return nil, fmt.Errorf("%s was empty", resource)
	} else if buf[0] != 0xFF {
		return nil, fmt.Errorf("the first byte of %s was 0x%x, not 0xFF", resource, buf[0])
	}
	metadataLen := binary.BigEndian.Uint32(buf[1:])
	if metadataLen < 1 || metadataLen > 65000 {
		return nil, fmt.Errorf("invalid metadata length for %s: %d", resource, metadataLen)
	}
	buf = buf[1+4+metadataLen:]

	magic := binary.BigEndian.Uint32(buf)
	const pbufMagic = 1346524486 // 4 bytes: "PBUF"
	if magic != pbufMagic {
		return nil, fmt.Errorf("invalid magic number for %s: %d", resource, magic)
	}
	buf = buf[4:]
	return buf, nil

}

func buildZKHosts(hosts string, zkport int) []string {
	hostSlice := strings.Split(hosts, ",")

	addrs := []string{}
	for _, host := range hostSlice {
		addrs = append(addrs, fmt.Sprintf("%v:%v", host, zkport))
	}
	return addrs
}

func (c *ZookeeperClient) Create(path string, data []byte, flags int32) error {
	_, err := c.conn.Create(path, data, flags, zk.WorldACL(zk.PermAll))
	return err
}

//Update data of give path, if not exist create one
func (c *ZookeeperClient) CreateOrUpdate(path string, data []byte, flags int32) error {
	err := c.CreateRecursive(path, data, flags)
	if err != nil && err == zk.ErrNodeExists {
		err = c.Set(path, data)
	}
	return err
}

//recursive create a node
func (c *ZookeeperClient) CreateRecursive(zkPath string, data []byte, flags int32) error {
	err := c.Create(zkPath, data, flags)
	if err == zk.ErrNoNode {
		err = c.CreateRecursive(path.Dir(zkPath), []byte{}, flags)
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
		err = c.Create(zkPath, data, flags)
	}
	return err
}

// recursive create a node, if it exists then omit
func (c *ZookeeperClient) CreateRecursiveIgnoreExist(path string, data []byte, flags int32) (err error) {
	if err = c.CreateRecursive(path, data, flags); err == zk.ErrNodeExists {
		err = nil
	}
	return err
}

//Delete a node by path.
func (c *ZookeeperClient) Delete(path string) error {
	return c.conn.Delete(path, defaultVersion)
}

//递归删除
func (c *ZookeeperClient) DeleteRecursive(zkPath string) error {
	err := c.Delete(zkPath)
	if err == nil {
		return nil
	}
	if err != zk.ErrNotEmpty {
		return err
	}

	children, _, err := c.conn.Children(zkPath)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err = c.DeleteRecursive(path.Join(zkPath, child)); err != nil {
			return err
		}
	}

	return c.Delete(zkPath)
}

// set data to given path
func (c *ZookeeperClient) Set(path string, data []byte) error {
	_, err := c.conn.Set(path, data, defaultVersion)
	return err
}

// test given path whether has sub-node
func (c *ZookeeperClient) HasChildren(path string) (bool, error) {
	children, _, err := c.conn.Children(path)
	if err != nil {
		return false, err
	}
	if len(children) == 0 {
		return false, nil
	}
	return true, nil
}
