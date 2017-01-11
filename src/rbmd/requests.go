package rbmd

import (
	"log"
	"strings"
	"encoding/json"
	// "bytes"

	"github.com/samuel/go-zookeeper/zk"
)

//RequestWatch watch for mount/umount requests
func (z ZooNode) RequestWatch(fqdn string) {
	requestsPath := strings.Join([]string{z.Path, "cluster", fqdn, "requests"}, "/")
	_, _, ch, err := z.Conn.ChildrenW(requestsPath)
	if err != nil {
		log.Print("[zk ERROR] ", err)
	}

	for {
		req := <-ch
		log.Print("[DEBUG] ch path ", req.Path)
		childrens, _, err := z.Conn.Children(requestsPath)
		if err != nil {
			break
		}
		for _, child := range childrens {
			p := strings.Join([]string{req.Path, child}, "/")
			request, _, err := z.Conn.Get(p)
			if err != nil {
				log.Print("[zk ERROR] ", err)
			}

			var r RBDDevice
			err = json.Unmarshal(request, &r)
			if err != nil {
				log.Print("[ERROR] ", err)
			}

			// 1) Map RBD 2) Mount FS
			if child == "mount" {
				std, err := r.MapDevice()
				if err != nil {
					z.RMR(p)
					z.Answer(fqdn, child, std, "FAIL")
					log.Print("[ERROR] Mapping error: ", string(std), err)
					break
				}
				err = r.MountFS(string(std))
				if err != nil {
					r.UnmapDevice()
					z.RMR(p)
					z.Answer(fqdn, child, std, "FAIL")
					log.Print("[ERROR] Mount filesystem error: ", err)
					break
				}
				z.Answer(fqdn, child, std, "OK")
			// 1) Unmount FS 2) Unmap RBD
			} else if child == "umount" {
				err := r.UnmountFS()
				if err != nil {
					z.RMR(p)
					z.Answer(fqdn, child, []byte("Failed umount device"), "FAIL")
					log.Print("[ERROR] Umount error: ", err)
					break
				}
				std, err := r.UnmapDevice()
				if err != nil {
					z.RMR(p)
					z.Answer(fqdn, child, std, "FAIL")
					log.Print("[ERROR] Unmapping error: ", string(std), err)
					break
				}
				z.Answer(fqdn, child, std, "OK")
			} else if child == "resolve" && z.GetLeader() == fqdn {
				if err := z.Resolve(fqdn); err != nil {
					log.Print("[ERROR] ", err)
					z.RMR(p)
				}
			} else {
				log.Print("[ERROR] Unknown request: ", child)
				z.RMR(p)
			}
			z.RMR(p)
		}
		break
	}
}

//Resolve delete node from quorum
func (z ZooNode) Resolve(fqdn string) error {
	resolvePath := strings.Join([]string{z.Path, "cluster", fqdn, "requests", "resolve"}, "/")

	r, _, err := z.Conn.Get(resolvePath)
	if err != nil {
		return err
	}
	
	var res Resolve
	if err := json.Unmarshal(r, &res); err != nil {
		return err
	}

	deadlyNodePath := strings.Join([]string{z.Path, "cluster", res.Node}, "/")
	z.RMR(resolvePath)
	z.RMR(deadlyNodePath)

	return nil
}

//ResolveRequest make request for resolve deadly.
func (z ZooNode) ResolveRequest(r Resolve) error {
	leader := z.GetLeader()
	resolvePath := strings.Join([]string{z.Path, "cluster", leader, "requests", "resolve"}, "/")

	jsReq, err := json.Marshal(r)
	if err != nil {
		return err
	}

	_, err = z.Conn.Create(resolvePath, jsReq, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		_, err := z.Conn.Set(resolvePath, jsReq, -1)
		if err != nil {
			log.Print("[zk ERROR] ", err)
			return err
		}
	}
	
	return nil
}

//Answer make answer
func (z ZooNode) Answer(fqdn string, req string, stderr []byte, t string) {
	answerPath := strings.Join([]string{z.Path, "cluster", fqdn, "answers", req}, "/")

	answer := MountState{t, string(stderr)}
	answerJSON, err := json.Marshal(answer)
	if err != nil {
		log.Print("[ERROR] ", err)
	}

	_, err = z.Conn.Create(answerPath, answerJSON, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		log.Print("[zk ERROR] ", err)
		_, err := z.Conn.Set(answerPath, answerJSON, -1)
		if err != nil {
			log.Print("[zk ERROR] ", err)
		}
	}
}

//MountRequest create node with mount requests from API
func (z ZooNode) MountRequest(r RBDDevice) error {
	jsReq, err := json.Marshal(r)
	if err != nil {
		return err
	}

	requestsPath := strings.Join([]string{z.Path, "cluster", r.Node, "requests", "mount"}, "/")
	_, err = z.Conn.Create(requestsPath, jsReq, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		_, err := z.Conn.Set(requestsPath, jsReq, -1)
		if err != nil {
			log.Print("[zk ERROR] ", err)
			return err
		}
	}
	return nil
}

//UmountRequest create node with umount requests from API
// OMFG: Needs merge with MountRequest
func (z ZooNode) UmountRequest(r RBDDevice) error {
	jsReq, err := json.Marshal(r)
	if err != nil {
		return err
	}

	requestsPath := strings.Join([]string{z.Path, "cluster", r.Node, "requests", "umount"}, "/")
	_, err = z.Conn.Create(requestsPath, jsReq, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		_, err := z.Conn.Set(requestsPath, jsReq, -1)
		if err != nil {
			log.Print("[zk ERROR] ", err)
			return err
		}
	}
	return nil
}

//WatchAnswer watch for answer
func (z ZooNode) WatchAnswer(fqdn string, t string) MountState {
	answersPath := strings.Join([]string{z.Path, "cluster", fqdn, "answers"}, "/")
	log.Print("[DEBUG] ", answersPath)
	_, _, ch, err := z.Conn.ChildrenW(answersPath)
	if err != nil {
		log.Print("[zk ERROR] 107 ", err)
	}

	var ms MountState
	var p string
	
	for {
		ans := <-ch
		log.Print("[DEBUG] ch answer path ", ans.Path)
		childrens, _, err := z.Conn.Children(answersPath)
		if err != nil {
			ms.Message = "Zk Error"
			ms.State = "FAIL"
			break
		}
		for _, child := range childrens {
			p = strings.Join([]string{ans.Path, child}, "/")
			answer, _, err := z.Conn.Get(p)
			if err != nil {
				log.Print("[zk ERROR] ", err)
				ms.Message = "Zk Error"
				ms.State = "FAIL"
				break
			}

			var m MountState
			err = json.Unmarshal(answer, &m)
			if err != nil {
				log.Print("[ERROR] ", err)
				ms.Message = "Zk Error"
				ms.State = "FAIL"
				break
			}

			if child == t {
				ms = m
				break
			}
		}
		break
	}
	z.RMR(p)
	return ms
}
