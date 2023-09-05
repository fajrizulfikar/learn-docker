package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func createNewRoot() {
	os.MkdirAll("/home/fajri/simple_docker/bin", 0755)
	cmd := exec.Command("cp", "-v", "/bin/ssh", "/home/fajri/simple_docker/bin")
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func main() {
	createNewRoot()

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
	err = os.WriteFile("/sys/fs/cgroup/memory/mydocker/memory.limit_in_bytes", []byte("50000000"), 0700)
	if err != nil {
		panic(err)
	}

	// Add this process to the cgroup
	pid := strconv.Itoa(os.Getpid())
	err = os.WriteFile("/sys/fs/cgroup/memory/mydocker/cgroup.procs", []byte(pid), 0700)
	if err != nil {
		panic(err)
	}

	// Change root filesystem to an isolated filesystem
	if err := syscall.Chroot("/home/fajri/simple_docker"); err != nil {
		panic(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		panic("Failed to get current directory: " + err.Error())
	}
	fmt.Println("Current working directory:", wd)

	// Execute a shell within the new namespaces
	cmd := "/bin/sh"
	args := []string{"sh"}

	if err := syscall.Exec(cmd, args, os.Environ()); err != nil {
		panic(err)
	}
}
