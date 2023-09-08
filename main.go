package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
)

func createNewRoot() {
	err := godotenv.Load("local.env")
	if err != nil {
		panic("Failed to load .env file: " + err.Error())
	}

	targetDir := os.Getenv("NEW_ROOT_DIR")
	binary := "/bin/sh"

	err = os.MkdirAll(targetDir+"/bin", 0755)
	if err != nil {
		panic("Failed to create new root: " + err.Error())
	}

	fmt.Println("targetDir", targetDir)
	cmd := exec.Command("cp", "-v", binary, targetDir+"/bin")
	err = cmd.Run()
	if err != nil {
		panic("Failed to copy /bin/sh: " + err.Error())
	}

	// Get shared dependencies
	cmd = exec.Command("ldd", binary)
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic("Failed to get dependencies: " + err.Error())
	}

	libs := strings.FieldsFunc(string(output), func(r rune) bool {
		return r == '\n' || r == ' '
	})

	// Copy shared dependencies to new root
	for _, lib := range libs {
		lib = strings.TrimSpace(lib)
		if strings.HasPrefix(lib, "/") {
			cpCmd := exec.Command("cp", "--parents", lib, targetDir)
			if err := cpCmd.Run(); err != nil {
				panic("Failed to copy dependency: " + err.Error())
			}

			// Check if the dependency is a linker
			if strings.Contains(lib, "ld-linux") {
				// Ensure the directory exists
				err = os.MkdirAll(targetDir+"/lib64", 0755)
				if err != nil {
					panic("Failed to create lib64 directory: " + err.Error())
				}

				cpCmd = exec.Command("cp", lib, targetDir+"/lib64/")
				if err := cpCmd.Run(); err != nil {
					panic("Failed to copy linker: " + err.Error())
				}
			}
		}
	}

}

func main() {
	createNewRoot()

	// Create namespace
	syscall.Unshare(syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS)

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

	// Open shell in isolated system
	if err := syscall.Exec("/bin/sh", []string{"sh"}, os.Environ()); err != nil {
		panic("Failed to exec /bin/sh: " + err.Error())
	}
}
