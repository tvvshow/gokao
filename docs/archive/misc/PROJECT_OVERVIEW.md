# Gaokao Application Assistant - Project Overview

This document provides a high-level overview of the Gaokao Application Assistant project, its architecture, and key components.

## High-Level Architecture

The project is a "Gaokao Application Assistant" (高考志愿填报系统) that uses AI for college recommendations and career planning. It has a microservices architecture with the following components:

*   **Backend**: Go (Gin)
*   **Frontend**: Vue.js 3
*   **Databases**: PostgreSQL, Redis
*   **High-Performance Computing**: C++
*   **Infrastructure**: Docker, Kubernetes

## Services

The `docker-compose.yml` file defines the following services:

*   `postgres`: PostgreSQL database
*   `redis`: Redis cache
*   `data-service`: Go service for providing data on universities, majors, and admissions.
*   `api-gateway`: Go service that acts as the main entry point for all API requests.
*   `user-service`: Go service for managing user authentication and profiles.
*   `recommendation-service`: Go service that generates personalized college recommendations, likely using the C++ engine.
*   `frontend`: Vue.js 3 user-facing web application.

## Build and Run Commands

The `Makefile` provides the following common commands:

*   `make build`: Build all services and the frontend.
*   `docker-compose up -d`: Run the entire application in a development environment.

## Code Structure

### Backend

The backend logic is organized into microservices located in the `services/` directory. Each service is a self-contained Go application built with the Gin framework.

*   **`api-gateway`**: This service acts as a reverse proxy, routing requests to the appropriate backend services. It also handles cross-cutting concerns like rate limiting, caching (with Redis), metrics (with Prometheus), and security (JWT authentication).

*   **`user-service`**: This service is responsible for user-related functionality, including registration, login, token refresh, and profile management.

*   **`data-service`**: This service provides access to the core data of the application, including universities, majors, and admission data.

*   **`recommendation-service`**: This service is the heart of the AI-powered recommendation feature. It uses a C++ bridge (`cppbridge`) to communicate with a high-performance C++ engine.

### Frontend

The frontend is a Vue.js 3 application.

*   **Main Entry Point**: `frontend/src/main.ts`
*   **Root Component**: `frontend/src/App.vue`
*   **State Management**: Pinia
*   **Routing**: Vue Router
*   **UI Library**: Element Plus

### C++ Modules

The C++ code is used for performance-critical tasks and is integrated with the Go services using Cgo.

*   **Device Fingerprinting**: The `services/cpp-modules/device-fingerprint` directory contains a pre-compiled shared library for device fingerprinting. The `services/device-auth-service/internal/cpp/device_fingerprint.go` file provides a Go wrapper for this library.

*   **Recommendation Engine**: The `services/recommendation-service/pkg/cppbridge` directory contains the bridge between the Go recommendation service and the C++ recommendation engine.
