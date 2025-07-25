# 🚀 Loex - Local Development Environment Manager

Loex is a powerful CLI tool for managing local development environments. Easily start, stop, and manage Frontend, Backend, and Database services across multiple projects with simple commands.

## ✨ Features

- 🔄 **Multi-Project Management**: Register and manage multiple projects
- 🚀 **One-Command Launch**: Start all services with a single command
- 🔍 **Auto-Detection**: Automatically detects service types and suggests commands
- 🎯 **Service Isolation**: Manage frontend, backend, and database services separately
- 📊 **Process Monitoring**: Track running services with PID management
- 📝 **Comprehensive Logging**: Separate logs for each service
- 🧙‍♂️ **Interactive Setup**: Wizard-guided project configuration

## 🛠 Installation

```bash
brew tap kjunh972/loex && brew install loex
```

For other installation methods, see [GitHub Releases](https://github.com/kjunh972/loex/releases).

### 📦 Updating

```bash
# Update to latest version
loex update
```

## 🚀 Quick Start

### 1. Initialize a Project

```bash
loex init myproject
```

### 2. Configure Services

```bash
# Auto-detect services (recommended)
cd /path/to/your/project
loex config detect myproject

# Interactive wizard
loex config wizard myproject
```

### 3. Start All Services

```bash
loex start myproject
```

### 4. Check Status

```bash
# Check all services status
loex status myproject

# View detailed project info with service status
loex list myproject
```

### 5. Stop Services

```bash
loex stop myproject
```

## 📚 Command Reference

### 📋 현재 지원하는 전체 명령어

**기본 관리:**
- `loex init [project]` - 프로젝트 초기화
- `loex list` / `loex list [project]` - 프로젝트/서비스 목록
- `loex remove [project]` - 프로젝트 삭제
- `loex rename [old] [new]` - 프로젝트 이름 변경

**서비스 설정:**
- `loex config detect [project]` - 자동 감지 (권장)
- `loex config wizard [project]` - 대화형 설정
- `loex config [project] [service] [command]` - 수동 설정
- `loex config edit [project] [service]` - 기존 설정 수정 
- `loex config delete [project] [service]` - 서비스 삭제 

**서비스 실행:**
- `loex start [project]` - 모든 서비스 시작
- `loex start [project] [service]` - 개별 서비스 시작
- `loex stop [project]` - 모든 서비스 중지
- `loex stop [project] [service]` - 개별 서비스 중지
- `loex restart [project]` - 모든 서비스 재시작 
- `loex status [project]` - 서비스 상태 확인

**시스템:**
- `loex update` - 최신 버전으로 업데이트

### Project Management

```bash
# Initialize a new project
loex init [project-name]

# List all projects
loex list

# Show detailed project info with service status
loex list [project-name]

# Remove a project
loex remove [project-name]

# Rename a project
loex rename [old-name] [new-name]
```

### Service Configuration

```bash
# Auto-detect services in current directory (recommended)
loex config detect [project-name]

# Interactive configuration wizard
loex config wizard [project-name]

# Manual configuration
loex config [project-name] [service] [command]

# Edit existing service configuration 
loex config edit [project-name] [service]

# Delete service configuration 
loex config delete [project-name] [service]
```

### Service Management

```bash
# Start all services
loex start [project-name]

# Start specific service
loex start [project-name] [service-name]

# Examples:
loex start myapp              # Start all services
loex start myapp frontend     # Start only frontend
loex start myapp backend      # Start only backend
loex start myapp db           # Start only database

# Stop all services
loex stop [project-name]

# Stop specific service
loex stop [project-name] [service-name]

# Restart all services 
loex restart [project-name]

# Check service status
loex status [project-name]
```

## 🔍 Auto-Detection

Loex automatically detects common project types and suggests appropriate commands when using the `config detect` or `config wizard` commands.

**💡 Important**: Run the command from your project's root directory to enable auto-detection. Loex analyzes files in the current directory to suggest the best commands for each service type.

### Frontend Services
- **React**: `npm start` (detects `react` in package.json)
- **React Native**: `npx react-native start`
- **Vue.js**: `npm run dev` (detects `vue` in dependencies)
- **Angular**: `npm start` (detects `@angular/core`)
- **Next.js**: `npm run dev` (detects `next`)

### Backend Services
- **Go**: `go run main.go` or `go run .`
- **Java/Spring**: `mvn spring-boot:run` or `./gradlew bootRun`
- **Python/Django**: `python manage.py runserver`
- **Python/Flask**: `python app.py`
- **Rust**: `cargo run`
- **JAR files**: `java -jar [filename].jar`

### Database Services
- **Local MySQL**: `brew services start mysql` (macOS) or `sudo systemctl start mysql` (Linux)
- **Local PostgreSQL**: `brew services start postgresql` (macOS) or `sudo systemctl start postgresql` (Linux)
- **Docker Compose**: `docker-compose up -d`
- **Docker MySQL**: `docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password mysql:8.0`
- **Docker PostgreSQL**: `docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:15`


## 🔧 Configuration Examples

### Example 1: React + Spring Boot + Local MySQL

```bash
# Initialize project
loex init webapp

# Auto-detect services (recommended)
cd /path/to/your/project
loex config detect webapp

# Start everything
loex start webapp
```

### Example 2: E-commerce Project Setup

```bash
# Initialize and configure with wizard
loex init ecommerce
loex config wizard ecommerce

# Or use auto-detection (recommended)
cd /path/to/your/project
loex config detect ecommerce
```

### Example 3: Manual Configuration

```bash
# Initialize project
loex init myapp

# Configure services manually
loex config myapp frontend "npm start"
loex config myapp backend "./gradlew bootRun"  
loex config myapp db "brew services start mysql"

# Start everything
loex start myapp
```

## 📋 Additional Commands

### System Management

```bash
# Update loex to latest version
loex update

# Check version information
loex version
loex -v
```

### Service Status Display

When using `loex list [project]`, you'll see service status indicators:

- **running ●**: Service is currently running
- **stopped ○**: Service is stopped
- **unknown ?**: Status cannot be determined

Example output:
```
Project: myproject
Services: 3

  frontend: running ●
    Command: npm start
    Directory: /path/to/frontend

  backend: stopped ○
    Command: ./gradlew bootRun
    Directory: /path/to/backend
```