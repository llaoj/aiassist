package sysinfo

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	configDir   = ".aiassist"
	sysInfoFile = "sysinfo.json"
)

// SystemInfo holds system environment information
type SystemInfo struct {
	OS             string   `json:"os"`              // Operating system (linux, darwin, windows)
	OSName         string   `json:"os_name"`         // Distribution name (Ubuntu, macOS, etc.)
	OSVersion      string   `json:"os_version"`      // OS version
	Arch           string   `json:"arch"`            // Architecture (amd64, arm64, etc.)
	Shell          string   `json:"shell"`           // Shell name (bash, zsh, etc.)
	ShellVersion   string   `json:"shell_version"`   // Shell version
	Hostname       string   `json:"hostname"`        // Server hostname
	Kernel         string   `json:"kernel"`          // Kernel version
	PackageManager string   `json:"package_manager"` // Package manager (apt, yum, brew, etc.)
	InitSystem     string   `json:"init_system"`     // Init system (systemd, launchd, etc.)
	IsContainer    bool     `json:"is_container"`    // Running in container
	HasSudo        bool     `json:"has_sudo"`        // Has sudo access
	User           string   `json:"user"`            // Current user
	HomeDir        string   `json:"home_dir"`        // Home directory
	PythonVersion  string   `json:"python_version"`  // Python version if available
	AvailableTools []string `json:"available_tools"` // Common tools available (docker, git, curl, etc.)
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir), nil
}

func GetSysInfoPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, sysInfoFile), nil
}

// LoadOrCollect loads system info from cache or collects if not exists
func LoadOrCollect() (*SystemInfo, error) {
	path, err := GetSysInfoPath()
	if err != nil {
		return nil, err
	}

	// If file exists, load from cache
	if _, err := os.Stat(path); err == nil {
		return Load()
	}

	// File doesn't exist, collect and save
	return CollectAndSave()
}

func Load() (*SystemInfo, error) {
	path, err := GetSysInfoPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read sysinfo file: %w", err)
	}

	var info SystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse sysinfo file: %w", err)
	}

	return &info, nil
}

func Collect() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Get current user and home directory
	if user := os.Getenv("USER"); user != "" {
		info.User = user
	}
	if home, err := os.UserHomeDir(); err == nil {
		info.HomeDir = home
	}

	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// Check if running in container
	info.IsContainer = detectContainer()

	// Check sudo access
	info.HasSudo = checkSudoAccess()

	// Collect OS-specific information
	switch runtime.GOOS {
	case "linux":
		collectLinuxInfo(info)
	case "darwin":
		collectDarwinInfo(info)
	case "windows":
		collectWindowsInfo(info)
	}

	collectShellInfo(info)
	collectPythonVersion(info)
	collectAvailableTools(info)

	return info, nil
}

func CollectAndSave() (*SystemInfo, error) {
	info, err := Collect()
	if err != nil {
		return nil, err
	}

	if err := Save(info); err != nil {
		return nil, err
	}

	return info, nil
}

func Save(info *SystemInfo) error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create directory if not exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := GetSysInfoPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sysinfo: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write sysinfo file: %w", err)
	}

	return nil
}

func collectLinuxInfo(info *SystemInfo) {
	// Try to get distribution info from /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "NAME=") {
				info.OSName = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
			} else if strings.HasPrefix(line, "VERSION=") {
				info.OSVersion = strings.Trim(strings.TrimPrefix(line, "VERSION="), "\"")
			} else if strings.HasPrefix(line, "ID=") {
				distroID := strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
				// Detect package manager based on distro
				detectPackageManager(info, distroID)
			}
		}
	}

	if output, err := exec.Command("uname", "-r").Output(); err == nil {
		info.Kernel = strings.TrimSpace(string(output))
	}

	// Detect init system
	if _, err := exec.Command("systemctl", "--version").Output(); err == nil {
		info.InitSystem = "systemd"
	} else if _, err := os.Stat("/sbin/init"); err == nil {
		info.InitSystem = "sysvinit"
	}
}

func collectDarwinInfo(info *SystemInfo) {
	info.OSName = "macOS"
	info.PackageManager = "brew" // Default, will verify if installed
	info.InitSystem = "launchd"

	if output, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
		info.OSVersion = strings.TrimSpace(string(output))
	}

	if output, err := exec.Command("uname", "-r").Output(); err == nil {
		info.Kernel = strings.TrimSpace(string(output))
	}
}

func collectWindowsInfo(info *SystemInfo) {
	info.OSName = "Windows"

	if output, err := exec.Command("cmd", "/c", "ver").Output(); err == nil {
		info.OSVersion = strings.TrimSpace(string(output))
	}
}

func collectShellInfo(info *SystemInfo) {
	// Get shell from SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		info.Shell = filepath.Base(shell)

		// Try to get shell version
		if output, err := exec.Command(shell, "--version").Output(); err == nil {
			// Get first line of version output
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				info.ShellVersion = strings.TrimSpace(lines[0])
			}
		}
	}
}

// FormatAsContext formats system info as context string for LLM
func (s *SystemInfo) FormatAsContext() string {
	var sb strings.Builder
	sb.WriteString("[System Environment]\n")
	sb.WriteString(fmt.Sprintf("OS: %s (%s)\n", s.OSName, s.OS))
	if s.OSVersion != "" {
		sb.WriteString(fmt.Sprintf("Version: %s\n", s.OSVersion))
	}
	sb.WriteString(fmt.Sprintf("Architecture: %s\n", s.Arch))
	if s.User != "" {
		sb.WriteString(fmt.Sprintf("User: %s", s.User))
		if s.HasSudo {
			sb.WriteString(" (has sudo)")
		}
		sb.WriteString("\n")
	}
	if s.Shell != "" {
		sb.WriteString(fmt.Sprintf("Shell: %s", s.Shell))
		if s.ShellVersion != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", s.ShellVersion))
		}
		sb.WriteString("\n")
	}
	if s.PackageManager != "" {
		sb.WriteString(fmt.Sprintf("Package Manager: %s\n", s.PackageManager))
	}
	if s.InitSystem != "" {
		sb.WriteString(fmt.Sprintf("Init System: %s\n", s.InitSystem))
	}
	if s.PythonVersion != "" {
		sb.WriteString(fmt.Sprintf("Python: %s\n", s.PythonVersion))
	}
	if s.IsContainer {
		sb.WriteString("Environment: Container\n")
	}
	if len(s.AvailableTools) > 0 {
		sb.WriteString(fmt.Sprintf("Available Tools: %s\n", strings.Join(s.AvailableTools, ", ")))
	}
	if s.Kernel != "" {
		sb.WriteString(fmt.Sprintf("Kernel: %s\n", s.Kernel))
	}
	if s.Hostname != "" {
		sb.WriteString(fmt.Sprintf("Hostname: %s\n", s.Hostname))
	}
	return sb.String()
}

func detectContainer() bool {
	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		if strings.Contains(content, "docker") || strings.Contains(content, "lxc") || strings.Contains(content, "kubepods") {
			return true
		}
	}

	return false
}

func checkSudoAccess() bool {
	// Try sudo -n true (non-interactive)
	cmd := exec.Command("sudo", "-n", "true")
	err := cmd.Run()
	return err == nil
}

func detectPackageManager(info *SystemInfo, distroID string) {
	switch distroID {
	case "ubuntu", "debian", "linuxmint":
		info.PackageManager = "apt"
	case "centos", "rhel", "fedora":
		if _, err := exec.LookPath("dnf"); err == nil {
			info.PackageManager = "dnf"
		} else {
			info.PackageManager = "yum"
		}
	case "arch", "manjaro":
		info.PackageManager = "pacman"
	case "alpine":
		info.PackageManager = "apk"
	case "opensuse":
		info.PackageManager = "zypper"
	}
}

func collectPythonVersion(info *SystemInfo) {
	// Try python3 first
	if output, err := exec.Command("python3", "--version").Output(); err == nil {
		info.PythonVersion = strings.TrimSpace(string(output))
		return
	}

	// Try python
	if output, err := exec.Command("python", "--version").Output(); err == nil {
		info.PythonVersion = strings.TrimSpace(string(output))
	}
}

func collectAvailableTools(info *SystemInfo) {
	tools := []string{"docker", "git", "curl", "wget", "kubectl", "helm", "terraform", "ansible"}
	available := []string{}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err == nil {
			available = append(available, tool)
		}
	}

	info.AvailableTools = available
}
