package rbmd

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

//ServerConf configuration of http api server
type ServerConf struct {
	Addr string
	Ws   string
}

//MountHandler /mount location
func (wr wrr) MountHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var m RBDDevice
	err := decoder.Decode(&m)
	log.Print("[DEBUG] ", m)
	var msE []byte
	if err != nil {
		msE, _ = json.Marshal(MountState{"FAIL", "JSON parse failure"})
		w.Write(msE)
		return
	}

	// var wCh chan MountState
	wCh := make(chan MountState, 1)
	go func() { wCh <- wr.z.WatchAnswer(m.Node, "mount") }()
	err = wr.z.MountRequest(m)
	if err != nil {
		w.Write(msE)
	}

	answerState := <-wCh
	log.Print(answerState)
	wr.z.RMR(strings.Join([]string{wr.z.Path, "cluster", wr.Fqdn, "answers", "mount"}, "/"))
	state, err := json.Marshal(answerState)
	if err != nil {
		log.Print("[ERROR] ", err)
		w.Write(msE)
	}
	w.Write(state)
}

//UmountHandler /umount location
func (wr wrr) UmountHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var m RBDDevice
	err := decoder.Decode(&m)
	log.Print("[DEBUG] ", m)
	var msE []byte
	if err != nil {
		msE, _ = json.Marshal(MountState{"FAIL", "JSON parse failure"})
		w.Write(msE)
		return
	}

	// var wCh chan MountState
	wCh := make(chan MountState, 1)
	go func() { wCh <- wr.z.WatchAnswer(m.Node, "umount") }()
	err = wr.z.UmountRequest(m)
	if err != nil {
		w.Write(msE)
	}

	answerState := <-wCh
	log.Print(answerState)
	wr.z.RMR(strings.Join([]string{wr.z.Path, "cluster", wr.Fqdn, "answers", "umount"}, "/"))
	state, err := json.Marshal(answerState)
	if err != nil {
		log.Print("[ERROR] ", err)
		w.Write(msE)
	}
	w.Write(state)
}

//Resolve resolve request
type Resolve struct {
	Node string `json:"node"`
}

//ResolveHandler resolve `deadly.` state. /resolve location
func (wr wrr) ResoleHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var res Resolve
	err := decoder.Decode(&res)
	var msE []byte
	if err != nil {
		msE, _ = json.Marshal(MountState{"FAIL", "JSON parse failure"})
		w.WriteHeader(500)
		w.Write(msE)
		return
	}

	if err := wr.z.ResolveRequest(res); err != nil {
		w.WriteHeader(500)
	}

	w.WriteHeader(200)
}

//wrr API
type wrr struct {
	Fqdn string
	z    ZooNode
}

//ServeHTTP start http server
func (s ServerConf) ServeHTTP(z ZooNode, fqdn string) {

	// Return JSON of full quorum status
	http.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write(z.GetState())
	})

	// Return string with quorum health check result
	http.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(z.GetQuorumHealth()))
	})

	// Return JSON of node description
	http.HandleFunc("/v1/node", func(w http.ResponseWriter, r *http.Request) {
		n := GetNodeState(fqdn)
		state, err := json.Marshal(n)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(state)
		return
	})

	// Return JSON mertrics
	http.HandleFunc("/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		m, err := GetMetrics(z)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		state, err := json.Marshal(m)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(state)
	})

	wr := wrr{fqdn, z}

	// Mount volume. Accept JSON. Return JSON.
	http.HandleFunc("/v1/mount", wr.MountHandler)

	// Umount volume. Accept JSON. Return JSON.
	http.HandleFunc("/v1/umount", wr.UmountHandler)

	// Umount volume. Accept JSON. Return JSON.
	http.HandleFunc("/v1/resolve", wr.ResoleHandler)

	server := &http.Server{
		Addr: s.Addr,
	}
	log.Fatal(server.ListenAndServe())
}

//Writer ws
type Writer struct {
	Upgrader websocket.Upgrader
	z        ZooNode
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
