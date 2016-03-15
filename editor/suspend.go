// +build !windows

package editor

import "syscall"

func (editor *Editor) Suspend() {
	editor.screen.Suspend()
	pid := syscall.Getpid()
	syscall.Kill(pid, syscall.SIGSTOP)
	editor.screen.Open()
}
