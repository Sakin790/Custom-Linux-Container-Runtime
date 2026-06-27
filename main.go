
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"docker/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sudo ./mycontainer run <command>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		fmt.Println("Unknown command")
	}
}

func run() {
	fmt.Printf("Running [Parent] - PID: %d\n", os.Getpid())

	basePath, _ := os.Getwd()
	lowerDir := filepath.Join(basePath, "alpine-rootfs")


	utils.CheckAndSetupRootFS(lowerDir)

	
	containerID := fmt.Sprintf("container-%d", time.Now().UnixNano())


	defer utils.CleanupContainer(containerID)

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	
	cmd.Env = append(os.Environ(), fmt.Sprintf("CONTAINER_ID=%s", containerID))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running child process: %v\n", err)
		os.Exit(1)
	}
}

func child() {
	fmt.Printf("Running [Child] - PID: %d\n", os.Getpid())


	containerID := os.Getenv("CONTAINER_ID")
	if containerID == "" {
		fmt.Println("Container ID not found in environment")
		os.Exit(1)
	}

	basePath, _ := os.Getwd()
	lowerDir := filepath.Join(basePath, "alpine-rootfs")
	upperDir := filepath.Join(basePath, "containers", containerID, "upper")
	workDir := filepath.Join(basePath, "containers", containerID, "work")
	mergedDir := filepath.Join(basePath, "containers", containerID, "merged")

	utils.CreateContainerDirs(upperDir, workDir, mergedDir)
	utils.MountOverlayFS(lowerDir, upperDir, workDir, mergedDir)
	utils.IsolateAndPivotRoot(mergedDir)

	command := os.Args[2]
	args := os.Args[2:]
	cmd := exec.Command(command, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	

	cmd.Env = []string{
		"PATH=/bin:/sbin:/usr/bin:/usr/sbin",
		"TERM=xterm-256color", 
		
	}

	
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("\n👋 Container shell exited with status: %d\n", exitErr.ExitCode())
		} else {
			fmt.Printf("Error executing command inside container: %v\n", err)
			os.Exit(1)
		}
	}
}
