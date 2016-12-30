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
func Run(zoo Zk, s ServerConf) {
	connection, err := zoo.InitConnection()
	if err != nil {
		log.Fatal(err)
	}
	fqdn, err := os.Hostname()

	z := ZooNode{zoo.Path, connection}

	// Serve HTTP API
	go s.ServeHTTP(z, fqdn)

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

//GetState return cluster status
func (z ZooNode) GetState() []byte {
	quorumStatePath := strings.Join([]string{z.Path, "/log/quorum"}, "")

	stateJSON, _, err := z.Conn.Get(quorumStatePath)
	if err != nil {
		log.Fatal(err)
	}

	return stateJSON
}
