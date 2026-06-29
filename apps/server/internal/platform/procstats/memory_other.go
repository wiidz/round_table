//go:build !linux && !darwin

package procstats

import "os"

func processRSS(pid int) (int64, bool) {
	if pid != os.Getpid() {
		return 0, false
	}
	return int64(heapInUse()), true
}
