package process

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/internal/logger"
	"github.com/kjunh972/loex/pkg/models"
)

type Manager struct {
	config *config.Manager
	logger *logger.Manager
}

func NewManager(config *config.Manager, logger *logger.Manager) *Manager {
	return &Manager{
		config: config,
		logger: logger,
	}
}

func (m *Manager) StartService(projectName string, serviceType models.ServiceType) error {
	project, err := m.config.LoadProject(projectName)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	service, exists := project.Services[serviceType]
	if !exists {
		return fmt.Errorf("service %s not configured for project %s", serviceType, projectName)
	}

	if isRunning, _ := m.IsServiceRunning(projectName, serviceType); isRunning {
		return fmt.Errorf("service %s is already running for project %s", serviceType, projectName)
	}

	parts := strings.Fields(service.Command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command for service %s", serviceType)
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = service.Dir

	logFile, err := m.logger.GetLogFile(projectName, serviceType)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start service %s: %w", serviceType, err)
	}

	if err := m.savePID(projectName, serviceType, cmd.Process.Pid, service.Command); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to save PID: %w", err)
	}

	fmt.Printf("Started %s service for project '%s' (PID: %d)\n", serviceType, projectName, cmd.Process.Pid)
	
	time.Sleep(500 * time.Millisecond)
	if !m.isProcessRunning(cmd.Process.Pid) {
		logPath := filepath.Join(m.logger.GetLogsDir(projectName), fmt.Sprintf("%s.log", serviceType))
		fmt.Printf("Service '%s' failed to start (exited immediately)\n", serviceType)
		fmt.Printf("Check logs: %s\n", logPath)
	}
	
	return nil
}

func (m *Manager) StopService(projectName string, serviceType models.ServiceType) error {
	pids, err := m.config.LoadProjectPIDs(projectName)
	if err != nil {
		return fmt.Errorf("failed to load PIDs: %w", err)
	}

	processInfo, exists := pids.Services[serviceType]
	if !exists {
		return fmt.Errorf("no running process found for service %s", serviceType)
	}

	if !isProcessRunning(processInfo.PID) {
		delete(pids.Services, serviceType)
		m.config.SaveProjectPIDs(pids)
		return fmt.Errorf("process %d is not running", processInfo.PID)
	}

	pgid, err := syscall.Getpgid(processInfo.PID)
	if err != nil {
		process, err := os.FindProcess(processInfo.PID)
		if err != nil {
			return fmt.Errorf("failed to find process %d: %w", processInfo.PID, err)
		}
		if err := process.Signal(syscall.SIGTERM); err != nil {
			process.Kill()
		}
	} else {
		if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		}
	}

	time.Sleep(1 * time.Second)

	delete(pids.Services, serviceType)
	if err := m.config.SaveProjectPIDs(pids); err != nil {
		return fmt.Errorf("failed to update PID file: %w", err)
	}

	fmt.Printf("Stopped %s service for project '%s'\n", serviceType, projectName)
	return nil
}

func (m *Manager) StopAllServices(projectName string) error {
	pids, err := m.config.LoadProjectPIDs(projectName)
	if err != nil {
		return fmt.Errorf("failed to load PIDs: %w", err)
	}

	if len(pids.Services) == 0 {
		return fmt.Errorf("no running services found for project %s", projectName)
	}

	var errors []string
	for serviceType := range pids.Services {
		if err := m.StopService(projectName, serviceType); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", serviceType, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop some services: %s", strings.Join(errors, ", "))
	}

	return nil
}

func (m *Manager) GetServiceStatus(projectName string, serviceType models.ServiceType) (string, error) {
	project, err := m.config.LoadProject(projectName)
	if err == nil {
		if service, exists := project.Services[serviceType]; exists {
			if strings.Contains(service.Command, "brew services start") {
				return m.getBrewServiceStatus(service.Command)
			}
		}
	}

	pids, err := m.config.LoadProjectPIDs(projectName)
	if err != nil {
		return "unknown", fmt.Errorf("failed to load PIDs: %w", err)
	}

	processInfo, exists := pids.Services[serviceType]
	if !exists {
		return "stopped", nil
	}

	if isProcessRunning(processInfo.PID) {
		return "running", nil
	} else {
		delete(pids.Services, serviceType)
		m.config.SaveProjectPIDs(pids)
		return "stopped", nil
	}
}

func (m *Manager) IsServiceRunning(projectName string, serviceType models.ServiceType) (bool, error) {
	status, err := m.GetServiceStatus(projectName, serviceType)
	return status == "running", err
}

func (m *Manager) GetAllServicesStatus(projectName string) (map[models.ServiceType]string, error) {
	project, err := m.config.LoadProject(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	status := make(map[models.ServiceType]string)
	for serviceType := range project.Services {
		serviceStatus, err := m.GetServiceStatus(projectName, serviceType)
		if err != nil {
			status[serviceType] = "error"
		} else {
			status[serviceType] = serviceStatus
		}
	}

	return status, nil
}

func (m *Manager) savePID(projectName string, serviceType models.ServiceType, pid int, command string) error {
	pids, err := m.config.LoadProjectPIDs(projectName)
	if err != nil {
		return err
	}

	if pids.Services == nil {
		pids.Services = make(map[models.ServiceType]models.ProcessInfo)
	}

	pids.Services[serviceType] = models.ProcessInfo{
		PID:       pid,
		Command:   command,
		StartTime: time.Now(),
		Status:    "running",
	}

	return m.config.SaveProjectPIDs(pids)
}

func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix systems, Signal(0) can be used to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (m *Manager) GetProcessDetails(projectName string, serviceType models.ServiceType) (*models.ProcessInfo, error) {
	pids, err := m.config.LoadProjectPIDs(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to load PIDs: %w", err)
	}

	processInfo, exists := pids.Services[serviceType]
	if !exists {
		return nil, fmt.Errorf("no process info found for service %s", serviceType)
	}

	return &processInfo, nil
}

func (m *Manager) GetLogs(projectName string, serviceType models.ServiceType, lines int) ([]string, error) {
	logPath := filepath.Join(m.logger.GetLogsDir(projectName), fmt.Sprintf("%s.log", serviceType))
	
	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"No logs found"}, nil
		}
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var logLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		logLines = append(logLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	if lines > 0 && len(logLines) > lines {
		return logLines[len(logLines)-lines:], nil
	}

	return logLines, nil
}

func (m *Manager) isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (m *Manager) getBrewServiceStatus(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) < 4 {
		return "stopped", nil
	}
	
	serviceName := parts[3]
	
	cmd := exec.Command("brew", "services", "list")
	output, err := cmd.Output()
	if err != nil {
		return "stopped", nil
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == serviceName {
			if fields[1] == "started" {
				return "running", nil
			}
			return "stopped", nil
		}
	}
	
	return "stopped", nil
}