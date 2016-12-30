package rbmd

import (
	"net/http"
	"log"
	"encoding/json"
)

//ServerConf configuration of http api server
type ServerConf struct {
	Addr string
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
		w.Write([]byte("Not implemented yet."))
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
