package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/pkg/models"
)

type Manager struct {
	config *config.Manager
}

func NewManager(config *config.Manager) *Manager {
	return &Manager{config: config}
}

func (m *Manager) GetLogsDir(projectName string) string {
	return m.config.GetLogsPath(projectName)
}

func (m *Manager) GetLogFile(projectName string, serviceType models.ServiceType) (*os.File, error) {
	logsDir := m.GetLogsDir(projectName)
	
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	logPath := filepath.Join(logsDir, fmt.Sprintf("%s.log", serviceType))
	
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(file, "\n=== %s Service Started at %s ===\n", serviceType, timestamp)

	return file, nil
}

func (m *Manager) GetLogPath(projectName string, serviceType models.ServiceType) string {
	return filepath.Join(m.GetLogsDir(projectName), fmt.Sprintf("%s.log", serviceType))
}

func (m *Manager) ClearLogs(projectName string, serviceType models.ServiceType) error {
	logPath := m.GetLogPath(projectName, serviceType)
	
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return nil
	}

	return os.Truncate(logPath, 0)
}

func (m *Manager) ClearAllLogs(projectName string) error {
	logsDir := m.GetLogsDir(projectName)
	
	files, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read logs directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			logPath := filepath.Join(logsDir, file.Name())
			if err := os.Truncate(logPath, 0); err != nil {
				return fmt.Errorf("failed to clear log file %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}