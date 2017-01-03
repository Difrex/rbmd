package rbmd

import (
	// "syscall"
	"io/ioutil"
	"net"
	"log"
	"strings"
	"regexp"
	"time"
)


//ClusterStatus Quorum status struct
type ClusterStatus struct {
	Quorum []Node `json:"quorum"`
	Health string `json:"health"`
	Zk string     `json:"zk"`
}

//Node Node status struct
type Node struct {
	Node string     `json:"node"`
	IP IPs          `json:"ip"`
	Updated int64   `json:"updated"`
	Mounts []Mount  `json:"mounts"`
}

// Mount struct
type Mount struct {
	Mountpoint string `json:"mountpoint"`
	Mountopts string  `json:"mountopts"`
	Fstype string     `json:"fstype"`
	Pool string       `json:"pool"`
	Block string      `json:"block"`
}

//IPs IP addresses
type IPs struct {
	V4 []string `json:"v4"`
	V6 []string `json:"v6"`
}

//GetMounts Parse /proc/mounts and get RBD mounts
func GetMounts() []Mount {
	var mounts []Mount
	m, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range strings.Split(string(m), "\n") {
		mount := strings.Split(line, " ")
		match, err := regexp.MatchString("^(/dev/rbd).*$", mount[0])
		if err != nil {
			log.Print("[ERROR] ", err)
		}
		if match {
			p := strings.Split(mount[0], "/")
			pool := p[len(p) - 2]
			mounts = append(mounts, Mount{
				mount[1],
				mount[3],
				mount[2],
				pool,
				p[len(p) -1],
			})
		}
	}

	return mounts
}

//GetMyIPs Exclude 127.0.0.1 
func GetMyIPs() IPs {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print("[ERROR] ", err)
	}

	var v4 []string
	var v6 []string
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Print("[ERROR] ", err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
                ip = v.IP
			case *net.IPAddr:
                ip = v.IP
			}
			if ip.String() != "127.0.0.1" && ip.String() != "::1" {
				match, err := regexp.MatchString("^.*:.*$", ip.String())
				if err != nil {
					log.Print("[ERROR] ", err)
				}
				if match {
					v6 = append(v6, ip.String())
				} else {
					v4 = append(v4, ip.String())
				}
			}
		}
	}

	return IPs{
		v4,
		v6,
	}
}


//GetNodeState Return Node struct
func GetNodeState(fqdn string) Node {
	var n Node

	n.Node = fqdn
	n.IP = GetMyIPs()
	n.Updated = time.Now().Unix()
	n.Mounts = GetMounts()

	return n
}


//MountState status of mount/umount
type MountState struct {
	State string   `json:"state"`
	Message string `json:"message"`
}

//RBDDevice rbd block device struct
type RBDDevice struct {
	Node string       `json:"node"`
	Pool string       `json:"pool"`
	Block string      `json:"block"`
	Mountpoint string `json:"mountpoint"`
	Mountopts string  `json:"mountopts"`
	Fstype string     `json:"fstype"`
}

//MapDevice map rbd block device
func (r RBDDevice) MapDevice() error {
	// image := strings.Join([]string{r.Pool, r.Block}, "/")
	
	
	return nil
}

//UnmapDevice unmap rbd block device
func (r RBDDevice) UnmapDevice() error {
	return nil
}

//MountFS mount file system
func (r RBDDevice) MountFS() error {
	return nil
}

//UnmountFS unmount file system
func (r RBDDevice) UnmountFS() error {
	return nil
}
