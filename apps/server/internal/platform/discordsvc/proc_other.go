//go:build !linux

package discordsvc

import "syscall"

func discordChildProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
