# Project Overview: Gaokao College Application System

This project, named "Gaokao College Application System" (高考志愿填报系统), is a multi-service application designed to assist with college application processes. It offers features such as college recommendation, data querying, user management, and payment processing.

The system adopts a microservice architecture:
*   **Backend:** Developed with Go 1.25 and the Gin framework, comprising several microservices located in the `services/` directory.
*   **Frontend:** Built using Vue 3, TypeScript, and Vite, residing in the `frontend/` directory.
*   **Recommendation Engine:** Integrates C++ modules (`cpp-modules/`) via CGO for advanced recommendation capabilities.
*   **Data Storage:** Utilizes PostgreSQL for relational data and Redis for caching/session management.
*   **Deployment & Operations:** Managed with Docker Compose and Nginx.

## Directory Structure Highlights

*   `services/`: Contains individual Go microservices (e.g., `api-gateway`, `data-service`, `user-service`, `payment-service`, `recommendation-service`, `monitoring-service`).
*   `frontend/`: The Vue 3 web application.
*   `pkg/`: Shared Go modules and utilities used across backend services.
*   `cpp-modules/`: Source code for C++ modules integrated into the recommendation service.
*   `docker/`: Docker-related configurations and scripts.
*   `docs/`: Project documentation, architecture reports, and design documents.

## Building and Running

The project provides a `Makefile` for streamlined build and development workflows.

### Quick Start with Docker (Recommended)

1.  **Install Dependencies:**
    ```bash
    make deps
    ```
2.  **Start Services:**
    ```bash
    docker compose up -d
    ```
3.  **Check Status:**
    ```bash
    docker compose ps
    ```

**Default Ports:**
*   Frontend: `80`
*   API Gateway: `8080`
*   Data Service: `8082`
*   User Service: `8083`
*   Recommendation Service: `8084`
*   Payment Service: `8085`
*   Monitoring Service: `8086`
*   PostgreSQL: `5433`
*   Redis: `6380`

### Local Development

*   **Build Backend Services:**
    ```bash
    make build-go
    ```
*   **Run Frontend in Development Mode:**
    ```bash
    cd frontend && npm run dev
    ```
    *(Note: `go.work` is configured for shared packages and service modules. Execute Go commands from the repository root for consistent dependency resolution.)*

## Testing and Quality Checks

*   **Full Build (Go + Frontend):**
    ```bash
    make build
    ```
*   **Run All Tests (Go + Frontend):**
    ```bash
    make test
    ```
*   **Run Go Tests Only:**
    ```bash
    make test-go
    ```
*   **Run Frontend Tests Only:**
    ```bash
    make test-frontend
    ```
*   **Frontend Linting:**
    ```bash
    cd frontend
    npm run lint
    ```
*   **Frontend Type Checking:**
    ```bash
    cd frontend
    npm run type-check
    ```

## API and Swagger Documentation

*   **API Gateway Entry:** `http://localhost:8080`
*   **Generate Swagger Documentation:** After modifying API comments in `services/api-gateway`, run the following to update Swagger docs:
    ```bash
    cd services/api-gateway
    go run github.com/swaggo/swag/cmd/swag@v1.8.12 init -g main.go -o docs --parseDependency --parseInternal
    ```

## Development Conventions

*   **Pre-commit Checks:** Before committing, it is recommended to run `make test` and `cd frontend && npm run lint && npm run type-check`.
*   **Commit Messages:** Follow Conventional Commits specification (e.g., `fix(scope): message`, `feat(scope): message`, `chore(scope): message`).
