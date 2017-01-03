package main

import (
	"flag"
	"rbmd"
	"strings"
	// "log"
)

var (
	zk string
	zkPath string
	tick int
	listen string
	ws string
)

func init() {
	flag.StringVar(&zk, "zk", "127.0.0.1:2181", "Zookeeper servers comma separated")
	flag.StringVar(&zkPath, "zkPath", "/rbmd", "Zookeeper path")
	flag.StringVar(&listen, "listen", "0.0.0.0:9076", "HTTP API listen address")
	flag.StringVar(&ws, "ws", "0.0.0.0:7690", "Websockets listen address")
	flag.IntVar(&tick, "tick", 5, "Tick time loop")
	flag.Parse()
}

func main() {
	config := rbmd.Zk{
		strings.Split(zk, ","),
		zkPath,
		tick,
	}
	s := rbmd.ServerConf{listen, ws}
	rbmd.Run(config, s)
}
