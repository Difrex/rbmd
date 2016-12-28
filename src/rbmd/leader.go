package rbmd

import (
	"strings"
	"log"
	"encoding/json"
	"time"
)

//Quorum quorum information
type Quorum struct {
	Quorum []Node `json:"quorum"`
	Leader string `json:"leader"`
	Health string `json:"health"`
}

//GetQuorumHealth return health check of cluster state
func (z ZooNode) GetQuorumHealth() string {
	helthPath := strings.Join([]string{z.Path, "/log/health"}, "")
	health, _, err := z.Conn.Get(helthPath)
	if err != nil {
		log.Print("[ERROR] ", err)
	}

	return string(health)
}

//SetQuorumHealth set current cluster health
func (z ZooNode) SetQuorumHealth(health string) {
	helthPath := strings.Join([]string{z.Path, "/log/health"}, "")
	z.EnsureZooPath("log/health")
	_, zoStat, _ := z.Conn.Get(helthPath)

	z.Conn.Set(helthPath, []byte(health), zoStat.Version)
}

//CheckAndSetHealth ...
func (z ZooNode) CheckAndSetHealth(childrens []string) {
	for _, child := range childrens {
		var childNode Node
		childStatePath := strings.Join([]string{z.Path, "/cluster/", child, "/state"}, "")
		childState, _, err := z.Conn.Get(childStatePath)
		if err != nil {
			log.Print("[ERROR] ", err)
		}
		json.Unmarshal(childState, &childNode)
		state, message := CheckMounts(childState)
		if !state {
			if childNode.Updated < (time.Now().Unix() - 9) {
				z.SetQuorumHealth(strings.Join(message, "\n"))
				return
			} 
			z.SetQuorumHealth("alive.")
			return
		}
	}

	currentHealth := strings.Split(z.GetQuorumHealth(), " ")
	if currentHealth[0] == "resizing." {
		for _, child := range childrens {
			if child == currentHealth[2] {
				z.SetQuorumHealth("alive.")
				return
			}
		}
	}
}

//UpdateQuorum set current cluster state
func (z ZooNode) UpdateQuorum(childrens []string) {
	quorumStatePath := strings.Join([]string{z.Path, "/log/quorum"}, "")
	z.EnsureZooPath("log/quorum")

	// Get nodes statuses
	var quorum Quorum
	for _, child := range childrens {
		var node Node
		childPath := strings.Join([]string{z.Path, "/cluster/", child, "/state"}, "")
		data, _, _ := z.Conn.Get(childPath)
		json.Unmarshal(data, &node)
		quorum.Quorum = append(quorum.Quorum, node)
	}

	quorum.Health = z.GetQuorumHealth()
	quorum.Leader = z.GetLeader()
	_, zoStat, _ := z.Conn.Get(quorumStatePath)
	q, err := json.Marshal(quorum)
	if err != nil {
		log.Print("[ERROR] ", err)
	}

	// Update
	log.Print("[DEBUG] Updating quorum")
	z.Conn.Set(quorumStatePath, q, zoStat.Version)
}

//GetLeader get current leader
func (z ZooNode) GetLeader() string {
	leaderPath := strings.Join([]string{z.Path, "/log/leader"}, "")
	leader, _, err := z.Conn.Get(leaderPath)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}

	return string(leader)
}

//SetLeader set current leader
func (z ZooNode) SetLeader(fqdn string) {
	leaderPath := strings.Join([]string{z.Path, "/log/leader"}, "")
	z.EnsureZooPath("log/leader")
	_, zoStat, err := z.Conn.Get(leaderPath)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}

	if z.GetLeader() != fqdn {
		log.Print("[DEBUG] I'm leader")
		z.Conn.Set(leaderPath, []byte(fqdn), zoStat.Version)
	}
}

//FindLeader return f.q.d.n of current leader
func (z ZooNode) FindLeader(fqdn string) {
	childrens, _, _, err := z.Conn.ChildrenW(strings.Join([]string{z.Path, "/cluster"}, ""))
	if err != nil {
		log.Print("[ERROR] ", err)
	}
	
	var state bool

	myState, _, _ := z.Conn.Get(strings.Join([]string{z.Path, "/cluster/", fqdn, "/state"}, ""))
	var node Node
	json.Unmarshal(myState, &node)

	state = z.CompareChilds(childrens, node)
	if state {
		z.SetLeader(fqdn)
	}

	z.CheckAndSetHealth(childrens)
	z.UpdateQuorum(childrens)
}

//CompareChilds return bool
// Compare childrens
func (z ZooNode) CompareChilds(childrens []string, node Node) (bool) {
	currentLeader := z.GetLeader()
	for _, child := range childrens {
		var childNode Node
		childStatePath := strings.Join([]string{z.Path, "/cluster/", child, "/state"}, "")
		childState, _, err := z.Conn.Get(childStatePath)
		if err != nil {
			log.Print("[ERROR] ", err)
		}
		json.Unmarshal(childState, &childNode)

		if childNode.Updated < (time.Now().Unix() - 9) {
			leader := z.GetLeader()
			if leader == child {
				z.SetLeader(node.Node)
			}
			log.Print("[DEBUG] ", childNode)
			childrens, _ := z.DestroyNode(child)
			z.UpdateQuorum(childrens)
			continue
		}
	
		// Compare updated time
		if node.Updated < childNode.Updated {
			return false
		}

		// if (time.Now().Unix() - 5) < childNode.Updated &&
		if childNode.Node == currentLeader {
			return false
		}
	}
	return true
}

//DestroyNode ...
// Delete node from quorum
func (z ZooNode) DestroyNode(fqdn string) ([]string, string) {
	log.Print("[WARNING] Deleting node ", fqdn, " from quorum!")

	childStatePath := strings.Join([]string{z.Path, "/cluster/", fqdn, "/state"}, "")
	childPath := strings.Join([]string{z.Path, "/cluster/", fqdn}, "")
	nodeStat, stateVersion, _ := z.Conn.Get(childStatePath)

	// Check node mounts
	mountStat, message := CheckMounts(nodeStat)
	if mountStat {
		_, childVersion, _ := z.Conn.Get(childPath)
		z.Conn.Delete(childStatePath, stateVersion.Version)
		z.Conn.Delete(childPath, childVersion.Version)
		z.SetQuorumHealth(strings.Join([]string{"resizing. node ", fqdn}, ""))
	}
	
	childrens, _, _, err := z.Conn.ChildrenW(strings.Join([]string{z.Path, "/cluster"}, ""))
	if err != nil {
		log.Print("[ERROR] ", err)
	}
	log.Print("[DEBUG] After destroy childs ", childrens)

	return childrens, strings.Join(message, "\n")
}


//CheckMounts ...
// Check mounts on down node
func CheckMounts(nodeStat []byte) (bool, []string) {
	var node Node

	err := json.Unmarshal(nodeStat, &node)
	if err != nil {
		log.Print("[ERROR] ", err)
	}

	var message []string
	if len(node.Mounts) > 0 {
		message = append(message, "deadly. ", "Reason: ", " NODE: ", node.Node)
		for _, mount := range  node.Mounts {
			message = append(message, " mountpoint: ", mount.Mountpoint, " block: ", mount.Block, " pool: ", mount.Pool)
		}
		return false, message
	}

	return true, message
}
