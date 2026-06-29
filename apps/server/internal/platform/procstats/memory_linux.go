//go:build linux

package procstats

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func processRSS(pid int) (int64, bool) {
	if rss, ok := linuxProcRSS(pid); ok {
		return rss, true
	}
	if pid == os.Getpid() {
		return rusageRSS()
	}
	return 0, false
}

func linuxProcRSS(pid int) (int64, bool) {
	path := fmt.Sprintf("/proc/%d/status", pid)
	f, err := os.Open(path)
	if err != nil {
		return 0, false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, false
		}
		kb, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil || kb <= 0 {
			return 0, false
		}
		return kb * 1024, true
	}
	return 0, false
}

func rusageRSS() (int64, bool) {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0, false
	}
	if ru.Maxrss <= 0 {
		return 0, false
	}
	return ru.Maxrss * 1024, true
}
