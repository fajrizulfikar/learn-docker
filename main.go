package main

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func createNewRoot() {
	err := os.MkdirAll("/home/fajri/simple_docker/bin", 0755)
	if err != nil {
		panic("Failed to create new root: " + err.Error())
	}
	cmd := exec.Command("cp", "-v", "/bin/sh", "/home/fajri/simple_docker/bin")
	err = cmd.Run()
	if err != nil {
		panic("Failed to copy /bin/sh: " + err.Error())
	}
}

func main() {
	createNewRoot()

	const CLONE_NEWUTS = 0x04000000
	const CLONE_NEWPID = 0x20000000
	const CLONE_NEWNS = 0x00020000
	const SYS_UNSHARE = 310

	_, _, errno := syscall.Syscall(uintptr(SYS_UNSHARE), uintptr(CLONE_NEWUTS|CLONE_NEWPID|CLONE_NEWNS), 0, 0)
	if errno != 0 {
		panic("Failed to unshare: " + errno.Error())
	}

	// Create cgroup directory
	err := os.Mkdir("/sys/fs/cgroup/memory/mydocker", 0755)
	if err != nil && !os.IsExist(err) {
		panic("Failed to create cgroup: " + err.Error())
	}

	// Create a new Memory cgroup and set its limit
	err = os.WriteFile("/sys/fs/cgroup/memory/mydocker/memory.limit_in_bytes", []byte("50000000"), 0700)
	if err != nil {
		panic("Failed to set memory limit: " + err.Error())
	}

	// Add this process to the cgroup
	pid := strconv.Itoa(os.Getpid())
	err = os.WriteFile("/sys/fs/cgroup/memory/mydocker/cgroup.procs", []byte(pid), 0700)
	if err != nil {
		panic("Failed to add process to cgroup: " + err.Error())
	}

	// Change root filesystem to an isolated filesystem
	if err := syscall.Chroot("/home/fajri/simple_docker"); err != nil {
		panic("Failed to chroot: " + err.Error())
	}

	// Execute a shell within the new namespaces
	if err := syscall.Exec("/bin/sh", []string{"sh"}, os.Environ()); err != nil {
		panic("Failed to exec /bin/sh: " + err.Error())
	}
}
