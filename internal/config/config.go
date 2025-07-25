package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kjunh972/loex/pkg/models"
)

const (
	ConfigDir     = ".loex"
	ProjectsDir   = "projects"
	PIDsDir       = "pids"
	LogsDir       = "logs"
)

type Manager struct {
	configPath string
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ConfigDir)
	
	if err := ensureDirectories(configPath); err != nil {
		return nil, fmt.Errorf("failed to create config directories: %w", err)
	}

	return &Manager{configPath: configPath}, nil
}

func ensureDirectories(basePath string) error {
	dirs := []string{
		basePath,
		filepath.Join(basePath, ProjectsDir),
		filepath.Join(basePath, PIDsDir),
		filepath.Join(basePath, LogsDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) GetProjectPath(name string) string {
	return filepath.Join(m.configPath, ProjectsDir, name+".json")
}

func (m *Manager) GetPIDPath(name string) string {
	return filepath.Join(m.configPath, PIDsDir, name+"-pids.json")
}

func (m *Manager) GetLogsPath(name string) string {
	return filepath.Join(m.configPath, LogsDir, name)
}

func (m *Manager) SaveProject(project *models.Project) error {
	project.Updated = time.Now()
	
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project: %w", err)
	}

	projectPath := m.GetProjectPath(project.Name)
	return os.WriteFile(projectPath, data, 0644)
}

func (m *Manager) LoadProject(name string) (*models.Project, error) {
	projectPath := m.GetProjectPath(name)
	
	data, err := os.ReadFile(projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("project '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read project file: %w", err)
	}

	var project models.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project: %w", err)
	}

	return &project, nil
}

func (m *Manager) ListProjects() ([]string, error) {
	projectsPath := filepath.Join(m.configPath, ProjectsDir)
	
	files, err := os.ReadDir(projectsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	var projects []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			name := file.Name()[:len(file.Name())-5]
			projects = append(projects, name)
		}
	}

	return projects, nil
}

func (m *Manager) ProjectExists(name string) bool {
	_, err := os.Stat(m.GetProjectPath(name))
	return err == nil
}

func (m *Manager) DeleteProject(name string) error {
	projectPath := m.GetProjectPath(name)
	pidPath := m.GetPIDPath(name)
	logsPath := m.GetLogsPath(name)

	if err := os.Remove(projectPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete project file: %w", err)
	}

	if err := os.Remove(pidPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete PID file: %w", err)
	}

	if err := os.RemoveAll(logsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete logs directory: %w", err)
	}

	return nil
}

func (m *Manager) RenameProject(oldName, newName string) error {
	if !m.ProjectExists(oldName) {
		return fmt.Errorf("project '%s' does not exist", oldName)
	}
	
	if m.ProjectExists(newName) {
		return fmt.Errorf("project '%s' already exists", newName)
	}

	project, err := m.LoadProject(oldName)
	if err != nil {
		return err
	}

	project.Name = newName
	if err := m.SaveProject(project); err != nil {
		return err
	}

	oldPIDPath := m.GetPIDPath(oldName)
	newPIDPath := m.GetPIDPath(newName)
	if _, err := os.Stat(oldPIDPath); err == nil {
		if err := os.Rename(oldPIDPath, newPIDPath); err != nil {
			return fmt.Errorf("failed to rename PID file: %w", err)
		}
	}

	oldLogsPath := m.GetLogsPath(oldName)
	newLogsPath := m.GetLogsPath(newName)
	if _, err := os.Stat(oldLogsPath); err == nil {
		if err := os.Rename(oldLogsPath, newLogsPath); err != nil {
			return fmt.Errorf("failed to rename logs directory: %w", err)
		}
	}

	return os.Remove(m.GetProjectPath(oldName))
}

func (m *Manager) SaveProjectPIDs(pids *models.ProjectPIDs) error {
	pids.Updated = time.Now()
	
	data, err := json.MarshalIndent(pids, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PIDs: %w", err)
	}

	pidPath := m.GetPIDPath(pids.ProjectName)
	return os.WriteFile(pidPath, data, 0644)
}

func (m *Manager) LoadProjectPIDs(name string) (*models.ProjectPIDs, error) {
	pidPath := m.GetPIDPath(name)
	
	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.ProjectPIDs{
				ProjectName: name,
				Services:    make(map[models.ServiceType]models.ProcessInfo),
			}, nil
		}
		return nil, fmt.Errorf("failed to read PID file: %w", err)
	}

	var pids models.ProjectPIDs
	if err := json.Unmarshal(data, &pids); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PIDs: %w", err)
	}

	return &pids, nil
}