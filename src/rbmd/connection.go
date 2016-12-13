package rbmd

import (
	"github.com/samuel/go-zookeeper/zk"
	"time"
	"log"
)


func (conf Zk) InitConnection() (*zk.Conn, error) {
	conn, _, err := zk.Connect(conf.Hosts, time.Second)
	if err != nil {
		log.Fatal(err)
	}

	return conn, err
}
