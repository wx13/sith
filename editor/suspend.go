// +build !windows

package editor

import "syscall"

// Suspend suspends the editor. Not available on windows.
func (editor *Editor) Suspend() {
	editor.screen.Suspend()
	pid := syscall.Getpid()
	syscall.Kill(pid, syscall.SIGSTOP)
	editor.screen.Open()
}
