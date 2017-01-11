package rbmd

import (
	"os"
	"runtime"
	"fmt"
)

//VersionShow show version and exit
func VersionShow() {
	fmt.Println("RBMD 0.0.3", runtime.Version(), runtime.GOARCH)
	os.Exit(1)
}
