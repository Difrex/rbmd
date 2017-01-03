package rbmd

import (
	"net/http"
	"log"
	"encoding/json"
	"time"
	
	"github.com/gorilla/websocket"
)

//ServerConf configuration of http api server
type ServerConf struct {
	Addr string
	Ws string
}

//ServeHTTP start http server
func (s ServerConf) ServeHTTP(z ZooNode, fqdn string) {

	// Return JSON of full quorum status
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write(z.GetState())
	})

	// Return string with quorum health check result
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(z.GetQuorumHealth()))
	})

	// Return JSON of node description
	http.HandleFunc("/node", func(w http.ResponseWriter, r *http.Request) {
		n := GetNodeState(fqdn)
		state, err := json.Marshal(n)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(state)
	})

	// Mount volume. Accept JSON. Return JSON.
	http.HandleFunc("/mount", func(w http.ResponseWriter, r *http.Request) {
		state, err := json.Marshal(MountState{"FAIL", "Not implemented yet"})
		if err != nil {
			log.Fatal(err)
		}
		w.Write(state)
	})

	// Umount volume. Accept JSON. Return JSON.
	http.HandleFunc("/umount", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Not implemented yet."))
	})

	server := &http.Server{
		Addr: s.Addr,
	}
	log.Fatal(server.ListenAndServe())
}


//Writer ws
type Writer struct {
	Upgrader websocket.Upgrader
	z ZooNode
}

//WriteStatusWs wrtite quorum status to websockets client
func (wr Writer) WriteStatusWs(w http.ResponseWriter, r *http.Request) {

	c, err := wr.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("[Ws ERROR] Upgrade: ", err)
	}

	mt, _, err := c.ReadMessage()
	if err != nil {
		log.Print("[Ws ERROR] Read error: ", err)
		// break
		return
	}
	
	go func() {
		for {
			err = c.WriteMessage(mt, wr.z.GetState())
			if err != nil {
				log.Print("[Ws ERROR] Write err: ", err)
				defer c.Close()
				break
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
}

//ServeWebSockets start websockets server
func (s ServerConf) ServeWebSockets(z ZooNode) {

	var wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Writer{wsUpgrader, z}.WriteStatusWs(w, r)
	})

	server := &http.Server{
		Addr: s.Ws,
	}
	log.Fatal(server.ListenAndServe())
}
