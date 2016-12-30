package rbmd

import (
	"log"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

//InitConnection Initialize Zookeeper connection
func (conf Zk) InitConnection() (*zk.Conn, error) {
	conn, _, err := zk.Connect(conf.Hosts, time.Second)
	if err != nil {
		log.Panic("[ERROR] ", err)
	}

	return conn, err
}

