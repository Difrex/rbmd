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
		log.Panic(err)
	}
	fqdn, err := os.Hostname()

	z := ZooNode{zoo.Path, connection}

	// Create Zk nodes tree
	z.EnsureZooPath("log/quorum")
	z.EnsureZooPath("log/health")
	z.EnsureZooPath("log/leader")
	z.EnsureZooPath(strings.Join([]string{"cluster/", fqdn, "/state"}, ""))

	// Serve HTTP API
	go s.ServeHTTP(z, fqdn)
	go s.ServeWebSockets(z)

	for {
		node, err := z.EnsureZooPath(strings.Join([]string{"cluster/", fqdn, "/state"}, ""))
		if err != nil {
			log.Panic("[ERROR] ", err)
		}
		z.UpdateState(node, fqdn)
		go z.FindLeader(fqdn)
		time.Sleep(time.Duration(zoo.Tick) * time.Second)
	}
}

//UpdateState -- update node status
func (z ZooNode) UpdateState(zkPath string, fqdn string) {
	z.EnsureZooPath(strings.Join([]string{"cluster/", fqdn, "/state"}, ""))
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


//jsonState HTTP API quorum status
type jsonState struct {
	Quorum map[string]Node `json:"quorum"`
    Health string          `json:"health"`
	Leader string          `json:"leader"`
}

//GetState return cluster status
func (z ZooNode) GetState() []byte {
	quorumStatePath := strings.Join([]string{z.Path, "/log/quorum"}, "")

	stateJSON, _, err := z.Conn.Get(quorumStatePath)
	if err != nil {
		log.Fatal(err)
	}

	var state Quorum
	json.Unmarshal(stateJSON, &state)

	node := make(map[string]Node)

	for _, n := range state.Quorum {
		node[n.Node] = n
	}

	retState := jsonState{node, state.Health, state.Leader}

	js, err := json.Marshal(retState)
	if err != nil {
		log.Fatal(err)
	}

	return js
}
