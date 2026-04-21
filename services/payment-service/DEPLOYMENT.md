# Payment and Membership Service - Deployment Guide

## Overview
This document provides instructions for deploying the payment and membership service.

## Prerequisites

### System Requirements
- Go 1.19 or higher
- PostgreSQL 13 or higher
- Redis 6.0 or higher
- Docker (optional, for containerized deployment)

### Dependencies
- All Go dependencies are managed through go.mod
- Database schema is automatically created on service startup

## Configuration

### Environment Variables
The service uses the following environment variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=payment_user
DB_PASSWORD=payment_password
DB_NAME=payment_db
DB_SSL_MODE=disable

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Server Configuration
SERVER_PORT=8084
SERVER_MODE=release
SERVER_READ_TIMEOUT=30
SERVER_WRITE_TIMEOUT=30

# JWT Configuration
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRATION=24h

# Payment Configuration
ALIPAY_APP_ID=your_alipay_app_id
ALIPAY_PRIVATE_KEY=your_alipay_private_key
ALIPAY_PUBLIC_KEY=alipay_public_key
WECHAT_APP_ID=your_wechat_app_id
WECHAT_MCH_ID=your_wechat_mch_id
WECHAT_API_KEY=your_wechat_api_key
```

### Configuration Files
The service can also be configured using a YAML configuration file:

```yaml
# config.yaml
database:
  host: localhost
  port: 5432
  user: payment_user
  password: payment_password
  database: payment_db
  ssl_mode: disable
  max_open_conns: 25
  max_idle_conns: 25

redis:
  addr: localhost:6379
  password: ""
  db: 0

server:
  port: 8084
  mode: release
  read_timeout: 30
  write_timeout: 30

jwt:
  secret: your_jwt_secret_key
  expiration: 24h

payment:
  alipay:
    app_id: your_alipay_app_id
    private_key: your_alipay_private_key
    public_key: alipay_public_key
  wechat:
    app_id: your_wechat_app_id
    mch_id: your_wechat_mch_id
    api_key: your_wechat_api_key
```

## Deployment Options

### 1. Direct Deployment

#### Build the Service
```bash
cd services/payment-service
go build -o payment-service
```

#### Run the Service
```bash
./payment-service
```

### 2. Docker Deployment

#### Build Docker Image
```bash
cd services/payment-service
docker build -t payment-service .
```

#### Run Docker Container
```bash
docker run -d \
  --name payment-service \
  -p 8084:8084 \
  -e DB_HOST=your_db_host \
  -e DB_USER=your_db_user \
  -e DB_PASSWORD=your_db_password \
  -e DB_NAME=your_db_name \
  -e REDIS_ADDR=your_redis_addr \
  payment-service
```

### 3. Docker Compose Deployment

#### docker-compose.yml
```yaml
version: '3.8'

services:
  payment-service:
    build: ./services/payment-service
    ports:
      - "8084:8084"
    environment:
      - DB_HOST=postgres
      - DB_USER=payment_user
      - DB_PASSWORD=payment_password
      - DB_NAME=payment_db
      - REDIS_ADDR=redis:6379
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  postgres:
    image: postgres:13
    environment:
      - POSTGRES_USER=payment_user
      - POSTGRES_PASSWORD=payment_password
      - POSTGRES_DB=payment_db
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped

  redis:
    image: redis:6
    ports:
      - "6379:6379"
    restart: unless-stopped

volumes:
  postgres_data:
```

#### Run with Docker Compose
```bash
docker-compose up -d
```

## Database Setup

The service automatically creates the required database schema on startup. However, you need to create the database and user manually:

```sql
CREATE DATABASE payment_db;
CREATE USER payment_user WITH PASSWORD 'payment_password';
GRANT ALL PRIVILEGES ON DATABASE payment_db TO payment_user;
```

## API Endpoints

### Health Check
- `GET /health` - Service health check

### Payment APIs
- `POST /api/v1/payments/create` - Create payment
- `POST /api/v1/payments/callback/:channel` - Payment callback
- `GET /api/v1/payments/query/:orderNo` - Query payment status
- `POST /api/v1/payments/refund` - Create refund
- `POST /api/v1/payments/close/:orderNo` - Close order
- `GET /api/v1/payments/channels` - Get supported payment channels

### Order APIs
- `POST /api/v1/orders/create` - Create order
- `GET /api/v1/orders/list` - List orders
- `GET /api/v1/orders/:orderNo` - Get order details
- `PUT /api/v1/orders/:orderNo/cancel` - Cancel order
- `GET /api/v1/orders/:orderNo/invoice` - Get invoice

### Membership APIs
- `GET /api/v1/membership/plans` - Get membership plans
- `POST /api/v1/membership/subscribe` - Subscribe to membership
- `GET /api/v1/membership/status` - Get membership status
- `POST /api/v1/membership/renew` - Renew membership
- `POST /api/v1/membership/cancel` - Cancel membership
- `GET /api/v1/membership/benefits` - Get membership benefits

## Monitoring and Logging

### Logging
The service outputs logs to stdout/stderr. For production deployments, consider using a log aggregation system like ELK or Fluentd.

### Metrics
The service exposes Prometheus metrics at `/metrics` endpoint.

### Health Checks
- Liveness probe: `/health`
- Readiness probe: `/health` (can be extended for more detailed checks)

## Security Considerations

### TLS/SSL
For production deployments, always use TLS/SSL encryption:

```bash
# In config.yaml
server:
  port: 443
  cert_file: /path/to/cert.pem
  key_file: /path/to/key.pem
```

### Rate Limiting
The service includes built-in rate limiting to prevent abuse.

### Authentication
All APIs require JWT authentication except for health check and payment callbacks.

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check database credentials and network connectivity
   - Verify database is running and accepting connections

2. **Redis Connection Failed**
   - Check Redis address and credentials
   - Verify Redis is running and accessible

3. **Payment Callback Not Working**
   - Check payment gateway configuration
   - Verify callback URL is publicly accessible

### Logs
Check service logs for detailed error information:

```bash
# Direct deployment
tail -f /var/log/payment-service.log

# Docker deployment
docker logs payment-service

# Docker Compose deployment
docker-compose logs payment-service
```

## Maintenance

### Backup
Regularly backup the PostgreSQL database:

```bash
pg_dump -h localhost -U payment_user payment_db > payment_db_backup.sql
```

### Updates
To update the service:

1. Pull the latest code
2. Build the new version
3. Stop the current service
4. Start the new version

### Scaling
The service is stateless and can be scaled horizontally by running multiple instances behind a load balancer.

## Support

For issues and support, please contact the development team or refer to the project documentation.