package main

import (
	"os"
	"strconv"
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

	// Create cgroup directory
	err := os.Mkdir("/sys/fs/cgroup/memory/mydocker", 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	// Create a new Memory cgroup and set its limit
	err = os.WriteFile("/sys/fs/cgroup/memory/mydocker/memory.limit_in_bytes", []byte("50000"), 0700)
	if err != nil {
		panic(err)
	}

	// Add this process to the cgroup
	pid := strconv.Itoa(os.Getpid())
	err = os.WriteFile("/sys/fs/cgroup/memory/mydocker/cgroup.procs", []byte(pid), 0700)
	if err != nil {
		panic(err)
	}

	// Execute a shell within the new namespaces
	cmd := "/bin/sh"
	args := []string{"sh"}

	if err := syscall.Exec(cmd, args, os.Environ()); err != nil {
		panic(err)
	}
}
