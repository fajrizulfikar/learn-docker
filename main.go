package main

import (
	"os"
	"syscall"
)

func main() {
	const CLONE_NEWUTS = 0x04000000
	const CLONE_NEWPID = 0x20000000
	const CLONE_NEWNS = 0x00020000
	const SYS_UNSHARE = 310

	if _, _, err := syscall.Syscall(uintptr(SYS_UNSHARE), uintptr(CLONE_NEWUTS|CLONE_NEWPID|CLONE_NEWNS), 0, 0); err != 0 {
		panic(err)
	}

	// Execute a shell within the new namespaces
	cmd := "/bin/sh"
	args := []string{"sh"}

	if err := syscall.Exec(cmd, args, os.Environ()); err != nil {
		panic(err)
	}
}
