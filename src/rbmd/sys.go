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


// Cluster status struct
type ClusterStatus struct {
	Quorum []Node
	Health string
	Zk string
}

// Node status struct
type Node struct {
	Node string
	Ip IPs
	Updated int
	Mounts []Mount
}

// Mount struct
type Mount struct {
	Mountpoint string
	Mountopts string
	Fstype string
	Pool string
	Block string
}

// IP addresses
type IPs struct {
	V4 []string
	V6 []string
}

// Parse /proc/mounts and get RBD mounts
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

// Exclude 127.0.0.1 
func GetMyIPs() IPs {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print("[ERROR] ", err)
	}

	var ipaddr IPs
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
				match, err := regexp.MatchString("^.*::.*$", ip.String())
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

	ipaddr.V4 = v4
	ipaddr.V6 = v6

	return ipaddr
}


// Return Node struct
func GetNodeState(fqdn string) Node {
	var n Node

	n.Node = fqdn
	n.Ip = GetMyIPs()
	n.Updated = int(time.Now().Unix())
	n.Mounts = GetMounts()

	return n
}
