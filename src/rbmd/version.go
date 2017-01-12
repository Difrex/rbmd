package rbmd

import (
	"fmt"
	"os"
	"runtime"
)

//VersionShow show version and exit
func VersionShow() {
	fmt.Println("RBMD 0.0.4", runtime.Version(), runtime.GOARCH)
	os.Exit(1)
}
