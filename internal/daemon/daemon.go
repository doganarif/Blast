package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// GetPIDPath returns the path to the daemon PID file
func GetPIDPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "blast", "daemon.pid"), nil
}

// GetLogPath returns the path to the daemon log file
func GetLogPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "blast", "daemon.log"), nil
}

// IsRunning checks if the daemon is currently running
func IsRunning() bool {
	pidPath, err := GetPIDPath()
	if err != nil {
		return false
	}

	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process is alive
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Start starts the daemon in the background
func Start() error {
	if IsRunning() {
		return nil // Already running
	}

	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get log file path
	logPath, err := GetLogPath()
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	// Start daemon process with platform-specific attributes
	cmd := exec.Command(exePath, "daemon")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	setProcAttributes(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Write PID file
	pidPath, err := GetPIDPath()
	if err != nil {
		return fmt.Errorf("failed to get PID path: %w", err)
	}

	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Wait a bit to ensure daemon started successfully
	time.Sleep(500 * time.Millisecond)

	if !IsRunning() {
		return fmt.Errorf("daemon failed to start")
	}

	return nil
}

// Stop stops the daemon
func Stop() error {
	pidPath, err := GetPIDPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Not running
		}
		return err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might be already dead
		os.Remove(pidPath)
		return nil
	}

	// Wait for process to exit
	for i := 0; i < 10; i++ {
		if !IsRunning() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Force kill if still running
	if IsRunning() {
		process.Kill()
	}

	// Remove PID file
	os.Remove(pidPath)

	return nil
}

// Reload sends a signal to the daemon to reload configuration
func Reload() error {
	pidPath, err := GetPIDPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(pidPath)
	if err != nil {
		return err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// Send SIGHUP to reload
	return process.Signal(syscall.SIGHUP)
}
