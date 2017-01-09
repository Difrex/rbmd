package rbmd

import (
	"syscall"
	"io/ioutil"
	"net"
	"log"
	"strings"
	"regexp"
	"time"
	"os/exec"
	"bytes"

	// "fmt"
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
	Image string      `json:"image"`
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
			pool, image := GetRBDPool(p[len(p) - 1])
			mounts = append(mounts, Mount{
				mount[1],
				mount[3],
				mount[2],
				pool,
				image,
				p[len(p) - 1],
			})
		}
	}

	return mounts
}

//GetRBDPool get pool from rbd showmapped
func GetRBDPool(device string) (string, string) {
	r := regexp.MustCompile(`^.*(\d+)$`)
	rbd := r.FindStringSubmatch(device)

	poolNamePath := strings.Join([]string{"/sys/bus/rbd/devices/", rbd[1], "/pool"}, "")
	imageNamePath := strings.Join([]string{"/sys/bus/rbd/devices/", rbd[1], "/name"}, "")

	pool, err := ioutil.ReadFile(poolNamePath)
	if err != nil {
		log.Fatal("[ERROR] Read failure ", err)
	}
	image, err := ioutil.ReadFile(imageNamePath)
	if err != nil {
		log.Fatal("[ERROR] Read failure ", err)
	}

	return string(pool), string(image)
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
	Image string      `json:"image"`
	Block string      `json:"block"`
	Mountpoint string `json:"mountpoint"`
	Mountopts string  `json:"mountopts"`
	Fstype string     `json:"fstype"`
}

//MapDevice map rbd block device
func (r RBDDevice) MapDevice() ([]byte, error) {
	image := strings.Join([]string{r.Pool, r.Image}, "/")
	log.Print("[DEBUG] Mapping ", image)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	
	cmd := exec.Command("rbd", "map", image)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return []byte(stderr.String()) , err
	}

	o := stdout.String()

	if strings.HasSuffix(o, "\n") {
        o = o[ :len(o) - 1]
    }
	
	return []byte(o), nil
}

//UnmapDevice unmap rbd block device
func (r RBDDevice) UnmapDevice() ([]byte, error) {
	log.Print("[DEBUG] Umapping ", r.Block)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	
	cmd := exec.Command("rbd", "unmap", strings.Join([]string{"/dev/", r.Block}, ""))

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return []byte(stderr.String()) , err
	}

	o := stdout.String()

	if strings.HasSuffix(o, "\n") {
        o = o[ :len(o) - 1]
    }
	
	return []byte(o), nil
}

//MountFS mount file system
func (r RBDDevice) MountFS(device string) error {
	err := syscall.Mount(device, r.Mountpoint, r.Fstype, 0, r.Mountopts)
	if err != nil {
		log.Print("[DEBUG] sys 207 ", err)
		return err
	}

	return nil
}

//UnmountFS unmount file system
func (r RBDDevice) UnmountFS() error {
	err := syscall.Unmount(r.Mountpoint, 0)
	if err != nil {
		log.Print("[DEBUG] sys 207 ", err)
		return err
	}

	return nil
}
