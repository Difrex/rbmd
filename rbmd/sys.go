package rbmd

import (
	"bytes"
	"io/ioutil"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

//ClusterStatus Quorum status struct
type ClusterStatus struct {
	Quorum []Node `json:"quorum"`
	Health string `json:"health"`
	Zk     string `json:"zk"`
}

//Node Node status struct
type Node struct {
	Node    string  `json:"node"`
	IP      IPs     `json:"ip"`
	Updated int64   `json:"updated"`
	Mounts  []Mount `json:"mounts"`
}

// Mount struct
type Mount struct {
	Mountpoint string `json:"mountpoint"`
	Mountopts  string `json:"mountopts"`
	Fstype     string `json:"fstype"`
	Pool       string `json:"pool"`
	Image      string `json:"image"`
	Block      string `json:"block"`
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
			log.Error(err.Error())
		}
		if match {
			p := strings.Split(mount[0], "/")
			pool, image := GetRBDPool(p[len(p)-1])
			mounts = append(mounts, Mount{
				Mountpoint: mount[1],
				Mountopts:  mount[3],
				Fstype:     mount[2],
				Pool:       pool,
				Image:      image,
				Block:      p[len(p)-1],
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

	p := strings.Trim(string(pool), "\n")
	i := strings.Trim(string(image), "\n")

	return p, i
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
	State   string `json:"state"`
	Message string `json:"message"`
}

//RBDDevice rbd block device struct
type RBDDevice struct {
	Node       string `json:"node"`
	Pool       string `json:"pool"`
	Image      string `json:"image"`
	Block      string `json:"block"`
	Mountpoint string `json:"mountpoint"`
	Mountopts  string `json:"mountopts"`
	Fstype     string `json:"fstype"`
}

//MapDevice map rbd block device
func (r RBDDevice) MapDevice() ([]byte, error) {
	image := strings.Join([]string{r.Pool, r.Image}, "/")
	log.Warn("[DEBUG] Mapping ", image)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("/usr/bin/rbd", "map", image)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Error(err.Error())
		return []byte(strings.Join([]string{stderr.String(), stdout.String()}, " ")), err
	}

	o := stdout.String()

	if strings.HasSuffix(o, "\n") {
		o = o[:len(o)-1]
	}

	return []byte(o), nil
}

//MountFS mount file system
func (r RBDDevice) MountFS(device string) error {
	err := syscall.Mount(device, r.Mountpoint, r.Fstype, ParseMountOpts(r.Mountopts), "")
	if err != nil {
		log.Error("Cant mount ", device, err.Error())
		return err
	}

	return nil
}

//ParseMountOpts parse RBDDevice.Mountopts. Return uintptr
func ParseMountOpts(mountopts string) uintptr {
	// Mount options map
	opts := make(map[string]uintptr)
	opts["ro"] = syscall.MS_RDONLY
	opts["posixacl"] = syscall.MS_POSIXACL
	opts["relatime"] = syscall.MS_RELATIME
	opts["noatime"] = syscall.MS_NOATIME
	opts["nosuid"] = syscall.MS_NOSUID
	opts["noexec"] = syscall.MS_NOEXEC
	opts["nodiratime"] = syscall.MS_NODIRATIME

	var msOpts uintptr
	if mountopts != "" {
		for _, o := range strings.Split(mountopts, ",") {
			msOpts = uintptr(msOpts | opts[o])
		}
		return msOpts
	}

	return 0
}

//UnmapDevice unmap rbd block device
func (r RBDDevice) UnmapDevice() ([]byte, error) {
	log.Print("[DEBUG] Umapping ", r.Block)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("/usr/bin/rbd", "unmap", strings.Join([]string{"/dev/", r.Block}, ""))

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return []byte(stderr.String()), err
	}

	o := stdout.String()

	if strings.HasSuffix(o, "\n") {
		o = o[:len(o)-2]
	}

	return []byte(o), nil
}

//UnmountFS unmount file system
func (r RBDDevice) UnmountFS() error {
	err := syscall.Unmount(r.Mountpoint, 0)
	log.Info("Try to umount ", r.Mountpoint)
	if err != nil {
		log.Error("Cant umount ", r.Mountpoint, err.Error())
		return err
	}

	return nil
}
