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
# Option A: Auto-detect services in current directory
cd /path/to/your/project
loex detect myproject

# Option B: Interactive wizard
loex config wizard myproject

# Option C: Manual configuration
loex config myproject frontend "npm start"
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
# Auto-detect services in current directory
loex detect [project-name]

# Interactive configuration wizard
loex config wizard [project-name]

# Set service manually (simplified syntax)
loex config [project] [service] [command]

# Examples:
loex config myapp frontend "npm start"
loex config myapp backend "./gradlew bootRun"
loex config myapp db "docker-compose up -d"
```

### Service Management

```bash
# Start all services
loex start [project-name]

# Start specific service
loex start [project-name] --service frontend

# Stop all services
loex stop [project-name]

# Stop specific service
loex stop [project-name] --service backend

# Check service status
loex status [project-name]
```

## 🔍 Auto-Detection

Loex automatically detects common project types and suggests appropriate commands when using the `detect` or `wizard` commands.

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

# Configure services
loex config webapp frontend "npm start"
loex config webapp backend "./gradlew bootRun"
loex config webapp db "brew services start mysql"

# Start everything
loex start webapp
```

### Example 2: React + Spring Boot + Docker MySQL

```bash
# Use auto-detection (run from project root directory)
cd /path/to/your/project
loex detect ecommerce

# Or use wizard for interactive setup
loex config wizard ecommerce

# Or configure manually
loex config ecommerce frontend "npm run dev"
loex config ecommerce backend "mvn spring-boot:run"
loex config ecommerce db "docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password mysql:8.0"
```

### Example 3: Local vs Docker Database Options

```bash
# Option A: Using Local Database
loex init shop-local
loex config shop-local frontend "npm run dev"
loex config shop-local backend "./gradlew bootRun"
loex config shop-local db "brew services start mysql"

# Option B: Using Docker Database
loex init shop-docker
loex config shop-docker frontend "npm run dev"
loex config shop-docker backend "./gradlew bootRun"
loex config shop-docker db "docker-compose up -d"

# Option C: Using PostgreSQL
loex init shop-postgres
loex config shop-postgres frontend "npm start"
loex config shop-postgres backend "mvn spring-boot:run"
loex config shop-postgres db "docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:15"

# Start any environment
loex start shop-local    # Local MySQL
loex start shop-docker   # Docker MySQL
loex start shop-postgres # Docker PostgreSQL
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

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request