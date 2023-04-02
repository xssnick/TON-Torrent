//go:build !windows

package daemon

import "syscall"

func daemonAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
