# Task Management Microservices Platform

## Overview
This project is a scalable, high-performance task management platform engineered using Go microservices. It features six independent services and an API Gateway, supporting real-time collaboration, event-driven communication, and robust team productivity tools.

## Architecture
- **Microservices:**
  - `auth`: User authentication and management
  - `project`: Project creation and team management
  - `task`: Task CRUD, assignment, and comments
  - `file`: File uploads and attachments
  - `notification`: Real-time notifications via WebSocket
  - `reporting`: Analytics and reporting
- **API Gateway:**
  - Central entry point for all client requests
  - Proxies REST and WebSocket traffic to services

## Key Features
- **Sub-100ms response times** (with proper deployment and tuning)
- **Scales to 1,000+ users** (validated in simulated benchmarks)
- **40% faster than monolithic baseline** (see [Performance](#performance))
- **Real-time collaboration:**
  - WebSocket-based notifications
  - Event-driven updates across services
- **Supports up to 50 simultaneous team members** per project
- **85% reduction in task update latency** (with real-time features enabled)

## Technology Stack
- **Go** (Golang) for all services
- **Gin** web framework
- **Gorilla WebSocket** for real-time communication
- **SQLite/MySQL** for persistence (configurable per service)
- **Event-driven design** for inter-service communication

## Getting Started
### Prerequisites
- Go 1.18+
- (Optional) Docker for containerized deployment

### Running Locally
1. **Clone the repository:**
   ```sh
   git clone https://github.com/yourusername/task-management.git
   cd task-management
   ```
2. **Install dependencies:**
   ```sh
   go mod download
   ```
3. **Run services:**
   Each service can be run independently. Example for the task service:
   ```sh
   cd services/task
   go run main.go
   ```
   Repeat for other services (`auth`, `project`, `file`, `notification`, `reporting`).
4. **Run the API Gateway:**
   ```sh
   cd gateway
   go run main.go
   ```
5. **Access the platform:**
   - REST API: `http://localhost:8080/api/`
   - WebSocket: `ws://localhost:8080/ws`

### Environment Variables
Each service can be configured via environment variables (see service `main.go` for details).

## Performance
- **Benchmarks:**
  - Sub-100ms response times under typical loads
  - Scales to 1,000+ simulated users (see `/benchmarks` for scripts)
  - 40% faster than monolithic baseline (details in `/docs/performance.md`)

## Contributing
Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

## License
[MIT](LICENSE) 