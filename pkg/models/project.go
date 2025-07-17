package models

import "time"

type ServiceType string

const (
	ServiceFrontend ServiceType = "frontend"
	ServiceBackend  ServiceType = "backend"
	ServiceDB       ServiceType = "db"
)

type Service struct {
	Type    ServiceType `json:"type"`
	Command string      `json:"command"`
	Dir     string      `json:"dir"`
	PID     int         `json:"pid,omitempty"`
	Status  string      `json:"status,omitempty"`
}

type Project struct {
	Name     string             `json:"name"`
	Services map[ServiceType]Service `json:"services"`
	Created  time.Time          `json:"created"`
	Updated  time.Time          `json:"updated"`
}

type ProcessInfo struct {
	PID       int       `json:"pid"`
	Command   string    `json:"command"`
	StartTime time.Time `json:"start_time"`
	Status    string    `json:"status"`
}

type ProjectPIDs struct {
	ProjectName string                 `json:"project_name"`
	Services    map[ServiceType]ProcessInfo `json:"services"`
	Updated     time.Time              `json:"updated"`
}