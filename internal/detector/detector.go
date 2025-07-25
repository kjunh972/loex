package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

	if frontend := d.detectFrontend(dir); frontend != nil {
		results = append(results, *frontend)
	}

	if backend := d.detectBackend(dir); backend != nil {
		results = append(results, *backend)
	}

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

	deps := extractDependencies(packageJSON)
	
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

	if contains(deps, "vue") {
		return &DetectionResult{
			Service:         models.ServiceFrontend,
			Command:         "npm run dev",
			DetectionReason: "Detected Vue.js project",
		}
	}

	if contains(deps, "@angular/core") {
		return &DetectionResult{
			Service:         models.ServiceFrontend,
			Command:         "npm start",
			DetectionReason: "Detected Angular project",
		}
	}

	if contains(deps, "next") {
		return &DetectionResult{
			Service:         models.ServiceFrontend,
			Command:         "npm run dev",
			DetectionReason: "Detected Next.js project",
		}
	}

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

	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		return &DetectionResult{
			Service:         models.ServiceBackend,
			Command:         "cargo run",
			DetectionReason: "Detected Rust project",
		}
	}

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
	if fileExists(filepath.Join(dir, "docker-compose.yml")) || fileExists(filepath.Join(dir, "docker-compose.yaml")) {
		return &DetectionResult{
			Service:         models.ServiceDB,
			Command:         "docker-compose up -d",
			DetectionReason: "Detected docker-compose.yml",
		}
	}

	if fileExists(filepath.Join(dir, "Dockerfile")) {
		return &DetectionResult{
			Service:         models.ServiceDB,
			Command:         "docker build -t local-db . && docker run -d local-db",
			DetectionReason: "Detected Dockerfile",
		}
	}

	if runtime.GOOS == "darwin" {
		if dbService := d.detectBrewDatabaseServices(); dbService != nil {
			return dbService
		}
	}

	if d.hasDBConfigFiles(dir) {
		fmt.Println("\nDatabase configuration detected but no database service found.")
		if runtime.GOOS == "darwin" {
			fmt.Println("To install MySQL:")
			fmt.Println("  brew install mysql")
		} else {
			fmt.Println("To install MySQL via Docker:")
			fmt.Println("  docker pull mysql:8.0")
		}
		fmt.Println("Then run 'loex config detect' again to register the database service.")
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

func (d *ServiceDetector) detectBrewDatabaseServices() *DetectionResult {
	cmd := exec.Command("brew", "services", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		
		serviceName := fields[0]
		
		if strings.Contains(serviceName, "mysql") {
			return &DetectionResult{
				Service:         models.ServiceDB,
				Command:         fmt.Sprintf("brew services start %s", serviceName),
				DetectionReason: fmt.Sprintf("Detected %s via Homebrew", serviceName),
			}
		}
		
		if strings.Contains(serviceName, "postgresql") || strings.Contains(serviceName, "postgres") {
			return &DetectionResult{
				Service:         models.ServiceDB,
				Command:         fmt.Sprintf("brew services start %s", serviceName),
				DetectionReason: fmt.Sprintf("Detected %s via Homebrew", serviceName),
			}
		}
	}
	
	return nil
}

func (d *ServiceDetector) hasDBConfigFiles(dir string) bool {
	configFiles := []string{
		"application.properties",
		"application.yml",
		"application.yaml",
		"database.yml",
		"database.yaml",
		"config/database.yml",
		"prisma/schema.prisma",
		"knexfile.js",
		"sequelize.js",
		"typeorm.config.js",
		"ormconfig.json",
	}
	
	for _, configFile := range configFiles {
		if fileExists(filepath.Join(dir, configFile)) {
			content, err := os.ReadFile(filepath.Join(dir, configFile))
			if err == nil && d.hasDBConfig(string(content)) {
				return true
			}
		}
	}
	
	return false
}

func (d *ServiceDetector) hasDBConfig(content string) bool {
	content = strings.ToLower(content)
	dbKeywords := []string{
		"jdbc:", "mysql", "postgres", "mongodb", "database_url",
		"db_host", "db_port", "datasource", "connection_string",
	}
	
	for _, keyword := range dbKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	
	return false
}