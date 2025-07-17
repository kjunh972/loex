package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kjunh972/loex/pkg/models"
)

type ServiceDetector struct{}

func New() *ServiceDetector {
	return &ServiceDetector{}
}

type DetectionResult struct {
	Service         models.ServiceType
	Command         string
	DetectionReason string
}

func (d *ServiceDetector) DetectServices(dir string) ([]DetectionResult, error) {
	var results []DetectionResult

	// Check for frontend services
	if frontend := d.detectFrontend(dir); frontend != nil {
		results = append(results, *frontend)
	}

	// Check for backend services
	if backend := d.detectBackend(dir); backend != nil {
		results = append(results, *backend)
	}

	// Check for database services
	if db := d.detectDatabase(dir); db != nil {
		results = append(results, *db)
	}

	return results, nil
}

func (d *ServiceDetector) detectFrontend(dir string) *DetectionResult {
	packageJSONPath := filepath.Join(dir, "package.json")
	
	if !fileExists(packageJSONPath) {
		return nil
	}

	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return nil
	}

	// Check dependencies
	deps := extractDependencies(packageJSON)
	
	// React detection
	if contains(deps, "react") {
		if contains(deps, "react-native") {
			return &DetectionResult{
				Service:         models.ServiceFrontend,
				Command:         "npx react-native start",
				DetectionReason: "Detected React Native project",
			}
		}
		return &DetectionResult{
			Service:         models.ServiceFrontend,
			Command:         "npm start",
			DetectionReason: "Detected React project",
		}
	}

	// Vue detection
	if contains(deps, "vue") {
		return &DetectionResult{
			Service:         models.ServiceFrontend,
			Command:         "npm run dev",
			DetectionReason: "Detected Vue.js project",
		}
	}

	// Angular detection
	if contains(deps, "@angular/core") {
		return &DetectionResult{
			Service:         models.ServiceFrontend,
			Command:         "npm start",
			DetectionReason: "Detected Angular project",
		}
	}

	// Next.js detection
	if contains(deps, "next") {
		return &DetectionResult{
			Service:         models.ServiceFrontend,
			Command:         "npm run dev",
			DetectionReason: "Detected Next.js project",
		}
	}

	// Generic Node.js frontend
	if scripts, ok := packageJSON["scripts"].(map[string]interface{}); ok {
		if _, hasStart := scripts["start"]; hasStart {
			return &DetectionResult{
				Service:         models.ServiceFrontend,
				Command:         "npm start",
				DetectionReason: "Detected Node.js project with start script",
			}
		}
		if _, hasDev := scripts["dev"]; hasDev {
			return &DetectionResult{
				Service:         models.ServiceFrontend,
				Command:         "npm run dev",
				DetectionReason: "Detected Node.js project with dev script",
			}
		}
	}

	return nil
}

func (d *ServiceDetector) detectBackend(dir string) *DetectionResult {
	// Go detection
	if fileExists(filepath.Join(dir, "go.mod")) {
		if fileExists(filepath.Join(dir, "main.go")) {
			return &DetectionResult{
				Service:         models.ServiceBackend,
				Command:         "go run main.go",
				DetectionReason: "Detected Go project with main.go",
			}
		}
		return &DetectionResult{
			Service:         models.ServiceBackend,
			Command:         "go run .",
			DetectionReason: "Detected Go project",
		}
	}

	// Java detection
	if fileExists(filepath.Join(dir, "pom.xml")) {
		return &DetectionResult{
			Service:         models.ServiceBackend,
			Command:         "mvn spring-boot:run",
			DetectionReason: "Detected Maven project",
		}
	}

	if fileExists(filepath.Join(dir, "build.gradle")) || fileExists(filepath.Join(dir, "build.gradle.kts")) {
		return &DetectionResult{
			Service:         models.ServiceBackend,
			Command:         "./gradlew bootRun",
			DetectionReason: "Detected Gradle project",
		}
	}

	// Python detection
	if fileExists(filepath.Join(dir, "requirements.txt")) || fileExists(filepath.Join(dir, "pyproject.toml")) {
		if fileExists(filepath.Join(dir, "manage.py")) {
			return &DetectionResult{
				Service:         models.ServiceBackend,
				Command:         "python manage.py runserver",
				DetectionReason: "Detected Django project",
			}
		}
		if fileExists(filepath.Join(dir, "app.py")) {
			return &DetectionResult{
				Service:         models.ServiceBackend,
				Command:         "python app.py",
				DetectionReason: "Detected Python Flask/FastAPI project",
			}
		}
	}

	// Rust detection
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		return &DetectionResult{
			Service:         models.ServiceBackend,
			Command:         "cargo run",
			DetectionReason: "Detected Rust project",
		}
	}

	// JAR file detection
	jarFiles, _ := filepath.Glob(filepath.Join(dir, "*.jar"))
	if len(jarFiles) > 0 {
		return &DetectionResult{
			Service:         models.ServiceBackend,
			Command:         fmt.Sprintf("java -jar %s", filepath.Base(jarFiles[0])),
			DetectionReason: "Detected JAR file",
		}
	}

	return nil
}

func (d *ServiceDetector) detectDatabase(dir string) *DetectionResult {
	// Docker Compose detection
	if fileExists(filepath.Join(dir, "docker-compose.yml")) || fileExists(filepath.Join(dir, "docker-compose.yaml")) {
		return &DetectionResult{
			Service:         models.ServiceDB,
			Command:         "docker-compose up -d",
			DetectionReason: "Detected docker-compose.yml",
		}
	}

	// Dockerfile detection
	if fileExists(filepath.Join(dir, "Dockerfile")) {
		return &DetectionResult{
			Service:         models.ServiceDB,
			Command:         "docker build -t local-db . && docker run -d local-db",
			DetectionReason: "Detected Dockerfile",
		}
	}

	// System database services
	if runtime.GOOS == "darwin" {
		return &DetectionResult{
			Service:         models.ServiceDB,
			Command:         "brew services start mysql",
			DetectionReason: "Default MySQL service for macOS",
		}
	} else if runtime.GOOS == "linux" {
		return &DetectionResult{
			Service:         models.ServiceDB,
			Command:         "sudo systemctl start mysql",
			DetectionReason: "Default MySQL service for Linux",
		}
	}

	return nil
}

func extractDependencies(packageJSON map[string]interface{}) []string {
	var deps []string
	
	if dependencies, ok := packageJSON["dependencies"].(map[string]interface{}); ok {
		for dep := range dependencies {
			deps = append(deps, dep)
		}
	}
	
	if devDependencies, ok := packageJSON["devDependencies"].(map[string]interface{}); ok {
		for dep := range devDependencies {
			deps = append(deps, dep)
		}
	}
	
	return deps
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}