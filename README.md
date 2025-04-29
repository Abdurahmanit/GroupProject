Bicycle Shop Microservices System
Overview
This project is a microservices-based system for a bicycle shop, designed to manage an online store's operations. It includes services for handling user accounts, product listings, reviews, orders, news, and admin tasks. The backend is built with Go, following clean architecture principles, and uses MongoDB for data storage, Redis for caching, and NATS for message queues.

Inter-service communication is handled via gRPC, while the API Gateway exposes REST endpoints over HTTP for external clients like the React frontend. Grafana is integrated for monitoring.

Architecture
The system consists of an API Gateway and six microservices:

User Service:
Manages user registration, login, logout, profile management, JWT authentication, and access control (user/admin roles).

Review Service:
Handles adding reviews for sellers and products, rating systems, and moderation features (report, hide, delete).

Listing Service:
Manages creation, updating, and deletion of bicycle listings, including search, filtering, photo uploads, and favorites/bookmarks.

Order Service:
Manages shopping cart operations, order placement, payment integration, order statuses, history, and PDF receipt generation.

News Service:
Handles publishing news articles, comments, and likes.

Admin Service:
Provides administrative functions like managing users, listings, reviews, news, and access rights.

The API Gateway acts as the entry point for external clients, accepting REST requests over HTTP from the frontend and translating them into gRPC calls to the appropriate microservices.

Each microservice follows clean architecture with layers for:

Entities (domain models)

Usecases (business logic)

Adapters (gRPC handlers)

Repositories (data access)

Project Structure
graphql
Копировать
Редактировать
GroupProject/
├── api-gateway/        # API Gateway for routing REST requests to gRPC services
├── user-service/       # User management service
├── review-service/     # Review and rating management service
├── listing-service/    # Product listing management service
├── order-service/      # Order and payment processing service
├── news-service/       # News publishing and interaction service
├── admin-service/      # Admin management service
├── frontend/           # React frontend for user interaction
├── scripts/            # Helper scripts for setup and running
├── monitoring/         # Grafana setup for monitoring
├── go.mod              # Go module file
└── README.md           # Project documentation
Example Microservice Structure: User Service
Each microservice follows a consistent structure, adhering to clean architecture principles. Below is the structure of the User Service as an example:

bash
Копировать
Редактировать
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
│   └── config/                       # Configuration
│       └── config.go
├── proto/                            # Protobuf definitions for gRPC
│   └── user.proto
├── tests/
│   ├── unit/
│   │   └── user_test.go
│   └── integration/
│       └── user_integration_test.go
└── go.mod                            # Service-specific dependencies
Each microservice has its own proto/ directory containing its gRPC definitions (e.g., user-service/proto/user.proto), ensuring independence and modularity.

Features
User Management: Register, login, logout, manage profiles, and enforce access control with JWT authentication.

Product Listings: Create, update, and delete bicycle listings with search, filtering, and photo uploads.

Reviews: Add and moderate reviews with ratings for products and sellers.

Orders: Add items to cart, place orders, process payments, track statuses, and generate PDF receipts.

News: Publish articles, allow comments, and track likes.

Admin Controls: Manage users, listings, reviews, news, and access rights.

Email Notifications: Send order confirmations via Google SMTP.

Monitoring: Use Grafana for tracing, metrics, and logs.

Ports
The frontend, API Gateway, and microservices each run on specific ports:


Component	Port	Description
Frontend (React)	3000	Serves all pages via client-side routing
API Gateway	8080	Accepts HTTP/REST requests from the frontend
User Service	50051	gRPC
Review Service	50052	gRPC
Listing Service	50053	gRPC
Order Service	50054	gRPC
News Service	50055	gRPC
Admin Service	50056	gRPC
Ports are configurable in each service’s internal/config/config.go.

Frontend Integration
The React frontend is a single-page application with multiple views (e.g., login, listings, orders) managed via client-side routing. It communicates with the API Gateway over HTTP using REST endpoints.

Example endpoints:

/api/user/register → User Service

/api/listings → Listing Service

/api/orders → Order Service

The API Gateway translates these into gRPC calls and returns JSON responses to the frontend.

Email Notifications
The Order Service sends email notifications (e.g., order confirmations) using Google SMTP.
Credentials should be configured in:
order-service/internal/config/config.go

Monitoring
Grafana is used to monitor system metrics, logs, and traces.
Default access: http://localhost:3000 (after setup)

Technologies Used
Backend: Go

Database: MongoDB

Caching: Redis (in-memory for sessions and search)

Messaging: NATS (for async communication like order notifications)

Inter-Service Communication: gRPC with Protocol Buffers

Frontend Communication: HTTP/REST via API Gateway

Frontend: React with Tailwind CSS

Monitoring: Grafana

PDF Generation: gofpdf

Notes
Ensure all services are running before accessing the frontend.

The API Gateway handles JWT authentication and request routing.

Redis is used for caching user sessions and search results.

PDF receipts are generated with the gofpdf library.
