package rbmd

import (
	// "github.com/samuel/go-zookeeper/zk"
	"log"
	"os"
	"strings"
	"time"
	_ "syscall"
	"encoding/json"
)

func Run(zoo Zk) {
	connection, err := zoo.InitConnection()
	if err != nil {
		log.Fatal(err)
	}
	fqdn, err := os.Hostname()

	z := ZooNode{zoo.Path, connection}
	node, err := z.EnsureZooPath(strings.Join([]string{"cluster/", fqdn, "/state"}, ""))

	for {
		go z.UpdateState(node, fqdn)
		go z.UpdateLeader()
		time.Sleep(time.Duration(zoo.Tick) * time.Second)
	}
}


func (z ZooNode) UpdateState(node string, fqdn string) {
	state := GetNodeState(fqdn)

	state_json, err := json.Marshal(state)
	if err != nil {
		log.Print("[ERROR] Failed encoding json ", err)
	}

	_, zo_stat, err := z.Conn.Get(node)
	if err != nil {
		log.Print("[ERROR] ", err)
	}

	log.Print("[DEBUG] ", "Updating state")
	zo_stat, err = z.Conn.Set(node, state_json, zo_stat.Version)
}


func (z ZooNode) UpdateLeader() {
	z.EnsureZooPath("log/leader")
}


func (z ZooNode) FindLeader() {
	children, _, _, err := z.Conn.ChildrenW(strings.Join([]string{z.Path, "/cluster"}, ""))
	if err != nil {
		log.Print("[ERROR] ", err)
	}

	for {
		for _, child := range children {
			log.Print("[DEBUG] ", child)
		}
	}
	
}

