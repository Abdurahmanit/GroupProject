
# Overview

This project is a microservices-based system for a bicycle shop, designed to manage an online store's operations. It includes services for handling user accounts, product listings, reviews, orders and news. The backend is built with Go, following clean architecture principles, and uses MongoDB for data storage, Redis for caching, and NATS for message queues. Inter-service communication is handled via gRPC, while the API Gateway exposes REST endpoints over HTTP for external clients like the React frontend. Grafana is integrated for monitoring.

## Architecture

The system consists of an API Gateway and five microservices:

- **User Service**: Manages user registration, login, logout, profile management, JWT authentication, and access control (user/admin roles).
- **Review Service**: Handles adding reviews for sellers and products, rating systems, and moderation features (report, hide, delete).
- **Listing Service**: Manages creation, updating, and deletion of bicycle listings, including search, filtering, photo uploads, and favorites/bookmarks, admin controls.
- **Order Service**: Manages shopping cart operations, order placement, payment integration, order statuses, history, and PDF receipt generation.
- **News Service**: Handles publishing news articles, comments, likes and admin controls.

The API Gateway acts as the entry point for external clients, accepting REST requests over HTTP from the frontend and translating them into gRPC calls to the appropriate microservices. The microservices communicate with each other exclusively via gRPC, ensuring efficient and type-safe interactions. Each microservice follows clean architecture with layers for entities (domain models), usecases (business logic), adapters (gRPC handlers), and repositories (data access).

## Project Structure

```bash
GroupProject/
├── api-gateway/      # API Gateway for routing REST requests to gRPC services
├── user-service/     # User management service
├── review-service/   # Review and rating management service
├── listing-service/  # Product listing management service
├── order-service/    # Order and payment processing service
├── news-service/     # News publishing and interaction service
├── frontend/         # React frontend for user interaction
├── scripts/          # Helper scripts for setup and running
├── monitoring/       # Grafana setup for monitoring
└── README.md         # Project documentation
```

## Example Microservice Structure: User Service

Each microservice follows a consistent structure, adhering to clean architecture principles. Below is the structure of the User Service as an example:

```bash
user-service/
├── cmd/
│   └── main.go                       # Entry point for the service
├── internal/
│   ├── entity/                       # Domain models
│   │   └── user.go
│   ├── usecase/                      # Application logic
│   │   └── user.go
│   ├── adapter/                      # Interface layer (gRPC handlers)
│   │   └── handler.go
│   ├── repository/                   # Data persistence (MongoDB)
│   │   └── user_repo.go
│   ├── config/                       # Configuration
│   │   └── config.go
│   └── proto/                        # Protobuf definitions for gRPC
│       └── user.proto
├── tests/
│   ├── unit/
│   │   └── user_test.go
│   └── integration/
│       └── user_integration_test.go
└── go.mod                            # Service-specific dependencies
```

Each microservice has its own `proto/` directory containing its gRPC definitions (e.g., `user-service/proto/user.proto`), ensuring independence and modularity.

## Features

- **User Management**: Register, login, logout, manage profiles, and enforce access control with JWT authentication.
- **Product Listings**: Create, update, and delete bicycle listings with search, filtering, and photo uploads.
- **Reviews**: Add and moderate reviews with ratings for products and sellers.
- **Orders**: Add items to cart, place orders, process payments, track statuses, and generate PDF receipts.
- **News**: Publish articles, allow comments, and track likes.
- **Admin Controls**: Manage users, listings, reviews, news, and access rights.
- **Monitoring**: Use Grafana for tracing, metrics, and logs.

## Ports

The frontend, API Gateway, and microservices each run on specific ports:

- Frontend (React): `3000`
- API Gateway: `8080`
- User Service: `50051` (gRPC)
- Listing Service: `50052` (gRPC)
- Review Service: `50053` (gRPC)
- Order Service: `50054` (gRPC)
- News Service: `50055` (gRPC)

Ports for microservices are configurable in each service’s `internal/config/config.go`.

## Frontend Integration

The React frontend is a single-page application with multiple pages (e.g., login, listings, orders) managed via client-side routing. It runs on a single port (`3000`) and communicates with the API Gateway using REST endpoints over HTTP. The API Gateway translates these requests into gRPC calls to the microservices and returns JSON responses to the frontend.

Example endpoints include:

- `/api/user/register` (User Service)
- `/api/listings` (Listing Service)
- `/api/orders` (Order Service)

## Monitoring

Grafana is set up for monitoring metrics, traces, and logs. Access Grafana at `http://localhost:3000` (default) after setting up the monitoring stack.

## Technologies Used

- **Backend**: Go
- **Database**: MongoDB
- **Caching**: Redis (in-memory store for caching user sessions and search results)
- **Message Queue**: NATS (lightweight messaging system for asynchronous communication, e.g., order notifications)
- **Inter-Service Communication**: gRPC (with Protocol Buffers)
- **Frontend Communication**: HTTP/REST (via API Gateway)
- **Frontend**: Common JS + CSS
- **Monitoring**: Grafana
- **PDF Generation**: gofpdf library for order receipts

## Notes

- Ensure all services are running before accessing the frontend.
- The API Gateway handles JWT authentication and request routing.
- Redis is used for caching user sessions and listing searches.
- PDF receipts for orders are generated using the gofpdf library.