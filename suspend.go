// +build !windows

package main

import "syscall"

func (editor *Editor) Suspend() {
	editor.screen.Close()
	pid := syscall.Getpid()
	syscall.Kill(pid, syscall.SIGSTOP)
	editor.screen.Open()
}
