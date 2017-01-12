package rbmd

import (
	"github.com/samuel/go-zookeeper/zk"
	"log"
	"strings"
	// "encoding/json"
)

//ZooNode zookeeper node
type ZooNode struct {
	Path string
	Conn *zk.Conn
	Zoo  Zk
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

//RMR remove Zk node recursive
func (z ZooNode) RMR(path string) {
	c, _, err := z.Conn.Children(path)
	if err != nil {
		log.Print("[zk ERROR] ", err)
	}
	log.Print("[WARNING] Trying delete ", path)
	if len(c) > 0 {
		for _, child := range c {
			childPath := strings.Join([]string{path, child}, "/")
			z.RMR(childPath)
		}
	}
	err = z.Conn.Delete(path, -1)
	if err != nil {
		log.Print("[zk ERROR] ", err)
	}
	log.Print("[WARNING] ", path, " deleted")
}
