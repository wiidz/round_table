//go:build darwin

package procstats

import (
	"os"
	"syscall"
)

func processRSS(pid int) (int64, bool) {
	if pid == os.Getpid() {
		return rusageRSS()
	}
	return psRSS(pid)
}

func rusageRSS() (int64, bool) {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0, false
	}
	if ru.Maxrss <= 0 {
		return 0, false
	}
	return ru.Maxrss, true
}
