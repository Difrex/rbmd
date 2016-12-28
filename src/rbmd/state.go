package rbmd

import (
	// "github.com/samuel/go-zookeeper/zk"
	"log"
	"os"
	"strings"
	"time"
	// "syscall"
	"encoding/json"
)

//Run -- start main loop
func Run(zoo Zk) {
	connection, err := zoo.InitConnection()
	if err != nil {
		log.Fatal(err)
	}
	fqdn, err := os.Hostname()

	z := ZooNode{zoo.Path, connection}

	for {
		node, err := z.EnsureZooPath(strings.Join([]string{"cluster/", fqdn, "/state"}, ""))
		if err != nil {
			log.Print("[ERROR] ", err)
		}
		go z.UpdateState(node, fqdn)
		go z.FindLeader(fqdn)
		time.Sleep(time.Duration(zoo.Tick) * time.Second)
	}
}

//UpdateState -- update node status
func (z ZooNode) UpdateState(zkPath string, fqdn string) {
	state := GetNodeState(fqdn)

	stateJSON, err := json.Marshal(state)
	if err != nil {
		log.Print("[ERROR] Failed encoding json ", err)
	}

	_, zoStat, err := z.Conn.Get(zkPath)
	if err != nil {
		log.Print("[ERROR] ", err)
	}

	log.Print("[DEBUG] ", "Updating state")
	zoStat, err = z.Conn.Set(zkPath, stateJSON, zoStat.Version)
	if err != nil {
		log.Print("[ERROR] ", err)
	}
}


