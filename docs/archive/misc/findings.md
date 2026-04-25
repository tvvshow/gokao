# Findings and Decisions

## Requirements

<!-- Captured from user request -->

## Research Findings

### High-Level Architecture

The project is a "Gaokao Application Assistant" (高考志愿填报系统) that uses AI for college recommendations and career planning. It has a microservices architecture with the following components:

*   **Backend**: Go (Gin) and Python (FastAPI)
*   **Frontend**: Vue.js 3
*   **Databases**: PostgreSQL, Redis
*   **High-Performance Computing**: C++
*   **Infrastructure**: Docker, Kubernetes

### Services

The `docker-compose.yml` file defines the following services:

*   `postgres`: PostgreSQL database
*   `redis`: Redis cache
*   `data-service`: Go service for providing data on universities, majors, and admissions.
*   `api-gateway`: Go service that acts as the main entry point for all API requests.
*   `user-service`: Go service for managing user authentication and profiles.
*   `recommendation-service`: Go service that generates personalized college recommendations, likely using the C++ engine.
*   `frontend`: Vue.js 3 user-facing web application.

### Build and Run Commands

The `Makefile` provides the following common commands:

*   `make build`: Build all services and the frontend.
*   `make build-go`: Build the Go services.
*   `make build-frontend`: Build the frontend application.
*   `make test`: Run all tests.
*   `make test-go`: Run tests for the Go services.
*   `make test-frontend`: Run tests for the frontend application.
*   `make deps`: Install and manage dependencies.
*   `make deps-go`: Install and manage dependencies for the Go services.
*   `make deps-frontend`: Install and manage dependencies for the frontend application.
*   `docker-compose up -d`: Run the entire application in a development environment.

### Backend Code Analysis

The backend logic is organized into microservices located in the `services/` directory. Each service is a self-contained Go application built with the Gin framework.

*   **`api-gateway`**: This service acts as a reverse proxy, routing requests to the appropriate backend services. It also handles cross-cutting concerns like rate limiting, caching (with Redis), metrics (with Prometheus), and security (JWT authentication). It uses a `ProxyManager` to manage the proxying logic.

*   **`user-service`**: This service is responsible for user-related functionality, including registration, login, token refresh, and profile management. It interacts with a PostgreSQL database and a Redis cache. It uses a `Permission` middleware for authorization.

*   **`data-service`**: This service provides access to the core data of the application, including universities, majors, and admission data. It offers a rich API for searching, filtering, and analyzing this data. It uses a PostgreSQL database and has a sophisticated caching mechanism.

*   **`recommendation-service`**: This service is the heart of the AI-powered recommendation feature. It uses a C++ bridge (`cppbridge`) to communicate with the high-performance C++ engine. It has two modes: a "simple_rule" engine and an "enhanced_rule" engine. It also has a data synchronization service to keep its data up-to-date.

### Frontend Code Analysis

The frontend is a Vue.js 3 application.

*   **Main Entry Point**: `frontend/src/main.ts`
*   **Root Component**: `frontend/src/App.vue`
*   **State Management**: Pinia
*   **Routing**: Vue Router
*   **UI Library**: Element Plus

The application is structured with a main `App.vue` component that sets up the layout with a header, footer, and a router view for rendering different pages. It uses Pinia for state management, with a `user` store to manage user authentication. It also uses Vue Router for client-side routing. The application uses Element Plus for UI components and has a dark mode implementation using `@vueuse/core`. The views are located in `frontend/src/views` and the components are in `frontend/src/components`.

### C++ Module Analysis

The C++ code is used for performance-critical tasks and is integrated with the Go services using Cgo.

*   **Device Fingerprinting**: The `services/cpp-modules/device-fingerprint` directory contains a pre-compiled shared library (`libdevice_fingerprint.so`) for device fingerprinting. The `services/device-auth-service/internal/cpp/device_fingerprint.go` file provides a Go wrapper for this library, allowing the `device-auth-service` to call the C++ functions for device fingerprinting.

*   **Recommendation Engine**: The `services/recommendation-service/pkg/cppbridge` directory contains the bridge between the Go recommendation service and the C++ recommendation engine. The `hybrid_bridge.go` file defines the `HybridRecommendationBridge` interface and a Cgo-based implementation that calls the C++ functions. The `enhanced_rule_bridge.go` file provides a pure Go implementation of the same interface, which is used when the C++ engine is not available. This allows for a flexible architecture where the core recommendation logic can be either in C++ for performance or in Go for easier development and debugging.

## Technical Decisions

<!-- Decisions made and their rationale -->

| Decision | Rationale |
| --- | --- |
| | |

## Issues Encountered

<!-- Errors and how they were resolved -->

| Issue | Resolution |
| --- | --- |
| | |

## Resources

<!-- URLs, file paths, API references -->

## Visual/Browser Findings

<!-- KEY: Update after every 2 view/browse actions -->
