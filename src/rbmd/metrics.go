package rbmd

import (
	"encoding/json"
	"runtime"
	"strings"
)

//Metrics metrics statistic
type Metrics struct {
	Goroutines  int   `json:"goroutines"`
	Nodes       int   `json:"nodes"`
	MountsTotal int   `json:"mountstotal"`
	Cgocall     int64 `json:"cgocall"`
}

// GetMetrics ...
func GetMetrics(z ZooNode) (Metrics, error) {
	var q Quorum
	var m Metrics

	curQuorum, _, err := z.Conn.Get(strings.Join([]string{z.Path, "log", "quorum"}, "/"))
	if err != nil {
		return m, err
	}

	err = json.Unmarshal(curQuorum, &q)
	if err != nil {
		return m, err
	}

	m.Nodes = len(q.Quorum)
	m.Goroutines = runtime.NumGoroutine()
	m.Cgocall = runtime.NumCgoCall()
	m.MountsTotal = GetTotalMounts(q.Quorum)

	return m, nil
}

// GetTotalMounts ...
func GetTotalMounts(n []Node) int {
	mounts := 0
	for _, node := range n {
		mounts = mounts + len(node.Mounts)
	}

	return mounts
}
