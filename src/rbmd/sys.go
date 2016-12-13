package rbmd

import (
	// "syscall"
	"io/ioutil"
	"log"
	"strings"
	"regexp"
)


// Node status struct
type Node struct {
	Node string
	Ip string
	Updated int
	Mounts []Mount
	Zk string
}

// Mount struct
type Mount struct {
	Mountpoint string
	Mountopts string
	Fstype string
	Pool string
	Block string
}


func GetMounts() ([]Mount, error) {
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

	return mounts, err
}
