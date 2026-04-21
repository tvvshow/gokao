# Payment and Membership Service

## Overview
The Payment and Membership Service is a comprehensive solution for handling payment processing and membership management in the 高考志愿填报助手 (College Application Assistant) platform. This service provides secure, scalable, and reliable payment processing with support for multiple payment channels and flexible membership plans.

## Features

### Payment Processing
- Support for multiple payment channels (Alipay, WeChat Pay, UnionPay)
- Secure payment processing with encryption
- Payment status tracking and management
- Refund processing
- Payment callback handling

### Membership Management
- Flexible membership plans with customizable features
- Subscription management
- Membership status tracking
- Automatic renewal options
- Membership benefits management

### Order Management
- Order creation and tracking
- Order status management
- Invoice generation
- Order history

### Security
- JWT-based authentication
- Input validation and sanitization
- Rate limiting
- Secure data encryption
- Audit logging

### Scalability
- Horizontal scaling support
- Database connection pooling
- Redis caching
- Load balancing support

## Architecture

### Technology Stack
- **Language**: Go
- **Database**: PostgreSQL
- **Cache**: Redis
- **API**: RESTful API
- **Authentication**: JWT
- **Payment Gateways**: Alipay, WeChat Pay, UnionPay

### Design Patterns
- **Repository Pattern**: For database operations
- **Service Layer**: For business logic
- **Handler Layer**: For HTTP request handling
- **Dependency Injection**: For loose coupling

### Database Schema
The service uses the following main tables:
- `payment_orders`: Payment order information
- `refund_records`: Refund records
- `membership_plans`: Membership plan definitions
- `user_memberships`: User membership information
- `payment_callbacks`: Payment callback logs
- `license_info`: License information

## API Documentation

### Base URL
```
http://localhost:8084/api/v1
```

### Authentication
All APIs (except health check and payment callbacks) require JWT authentication in the `Authorization` header:
```
Authorization: Bearer <your-jwt-token>
```

### Endpoints

#### Health Check
- `GET /health` - Service health check

#### Payment APIs
- `POST /payments/create` - Create payment
- `POST /payments/callback/{channel}` - Payment callback
- `GET /payments/query/{orderNo}` - Query payment status
- `POST /payments/refund` - Create refund
- `POST /payments/close/{orderNo}` - Close order
- `GET /payments/channels` - Get supported payment channels

#### Order APIs
- `POST /orders/create` - Create order
- `GET /orders/list` - List orders
- `GET /orders/{orderNo}` - Get order details
- `PUT /orders/{orderNo}/cancel` - Cancel order
- `GET /orders/{orderNo}/invoice` - Get invoice

#### Membership APIs
- `GET /membership/plans` - Get membership plans
- `POST /membership/subscribe` - Subscribe to membership
- `GET /membership/status` - Get membership status
- `POST /membership/renew` - Renew membership
- `POST /membership/cancel` - Cancel membership
- `GET /membership/benefits` - Get membership benefits

## Getting Started

### Prerequisites
- Go 1.19+
- PostgreSQL 13+
- Redis 6.0+

### Installation
1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Configure environment variables
4. Run the service: `go run main.go`

### Configuration
See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed configuration options.

## Testing

### Unit Tests
Run unit tests:
```bash
go test -v ./...
```

### Integration Tests
Run integration tests:
```bash
go test -v -tags=integration ./...
```

### End-to-End Tests
Run end-to-end tests:
```bash
go test -v -tags=e2e ./...
```

## Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

## Monitoring

### Health Checks
- Liveness probe: `/health`
- Readiness probe: `/health`

### Metrics
Prometheus metrics are available at `/metrics`.

### Logging
The service outputs structured JSON logs to stdout.

## Contributing

### Code Structure
```
internal/
├── models/          # Data models and database operations
├── services/        # Business logic
├── handlers/        # HTTP request handlers
├── middleware/      # HTTP middleware
├── database/        # Database initialization
├── config/          # Configuration
├── utils/           # Utility functions
└── adapters/        # Payment gateway adapters
```

### Development Workflow
1. Create a feature branch
2. Implement changes
3. Write tests
4. Run tests
5. Commit changes
6. Create pull request

### Code Standards
- Follow Go code conventions
- Write comprehensive tests
- Document public APIs
- Use meaningful commit messages

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support
For issues and support, please contact the development team or open an issue on the repository.