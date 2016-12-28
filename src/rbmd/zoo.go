package rbmd

import (
	"github.com/samuel/go-zookeeper/zk"
	"strings"
)

//ZooNode zookeeper node
type ZooNode struct {
	Path string
	Conn *zk.Conn
}

//EnsureZooPath create zookeeper path  
func (z ZooNode) EnsureZooPath(node string) (string, error) {
	flag := int32(0)
	acl := zk.WorldACL(zk.PermAll)

	zoopath := strings.Join([]string{z.Path, "/", node}, "")
	s := strings.Split(zoopath, "/")

	var p []string
	var fullnodepath string
	
	for i := 1; i < len(s); i++ {
		p = append(p, strings.Join([]string{"/", s[i]}, ""))
	}
	
	for i := 0; i < len(p); i++ {
		fullnodepath = strings.Join([]string{fullnodepath, p[i]}, "")
		z.Conn.Create(fullnodepath, []byte(""), flag, acl)
	}
	
	return fullnodepath, nil
}

