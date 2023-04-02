//go:build windows

package daemon

func daemonAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}
