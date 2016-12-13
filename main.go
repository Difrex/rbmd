package main

import (
	"flag"
	"rbmd"
	"strings"
	// "log"
)

var (
	zk string
	zk_path string
	tick int
)

func init() {
	flag.StringVar(&zk, "zk", "127.0.0.1:2181", "Zookeeper servers comma separated")
	flag.StringVar(&zk_path, "zk_path", "/rbmd", "Zookeeper path")
	flag.IntVar(&tick, "tick", 5, "Tick time loop")
	flag.Parse()
}

func main() {
	config := rbmd.Zk{
		strings.Split(zk, ","),
		zk_path,
		tick,
	}
	rbmd.Run(config)
}
