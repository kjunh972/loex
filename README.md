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

## 🚀 Quick Start

### 1. Initialize a Project

```bash
loex init myproject
```

### 2. Configure Services (Interactive)

```bash
# Navigate to your project directory first for auto-detection
cd /path/to/your/project
loex config wizard myproject
```

### 3. Start All Services

```bash
loex start myproject
```

### 4. Check Status

```bash
loex status myproject
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

# Remove a project
loex remove [project-name]

# Rename a project
loex rename [old-name] [new-name]
```

### Service Configuration

```bash
# Interactive configuration wizard
loex config wizard [project-name]

# Set service manually
loex config set [project] [service] [command] --dir [directory]

# Examples:
loex config set myapp frontend "npm start" --dir ./frontend
loex config set myapp backend "./gradlew bootRun" --dir ./backend
loex config set myapp db "docker-compose up -d" --dir .
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

Loex automatically detects common project types and suggests appropriate commands when using the `wizard` command.

**💡 Important**: Run the wizard command from your project's root directory to enable auto-detection. Loex analyzes files in the current directory to suggest the best commands for each service type.

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
loex config set webapp frontend "npm start" --dir ./frontend
loex config set webapp backend "./gradlew bootRun" --dir ./backend  
loex config set webapp db "brew services start mysql" --dir .

# Start everything
loex start webapp
```

### Example 2: React + Spring Boot + Docker MySQL

```bash
# Use wizard for interactive setup (run from project root directory)
cd /path/to/your/project
loex config wizard ecommerce

# Or configure manually
loex config set ecommerce frontend "npm run dev" --dir ./react-app
loex config set ecommerce backend "mvn spring-boot:run" --dir ./spring-api
loex config set ecommerce db "docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password mysql:8.0" --dir .
```

### Example 3: Local vs Docker Database Options

```bash
# Option A: Using Local Database
loex init shop-local
loex config set shop-local frontend "npm run dev" --dir ./frontend
loex config set shop-local backend "./gradlew bootRun" --dir ./backend
loex config set shop-local db "brew services start mysql" --dir .

# Option B: Using Docker Database
loex init shop-docker
loex config set shop-docker frontend "npm run dev" --dir ./frontend
loex config set shop-docker backend "./gradlew bootRun" --dir ./backend
loex config set shop-docker db "docker-compose up -d" --dir .

# Option C: Using PostgreSQL
loex init shop-postgres
loex config set shop-postgres frontend "npm start" --dir ./frontend
loex config set shop-postgres backend "mvn spring-boot:run" --dir ./backend
loex config set shop-postgres db "docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:15" --dir .

# Start any environment
loex start shop-local    # Local MySQL
loex start shop-docker   # Docker MySQL
loex start shop-postgres # Docker PostgreSQL
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request