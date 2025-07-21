# Subscriptions Aggregator API

A subscription management service built with Go. This API allows you to:
- Create and manage user subscriptions
- Calculate subscription costs for specific periods
- Handle subscription renewals and cancellations

## Features

- **Subscription Lifecycle**: Full CRUD operations for subscriptions
- **Cost Calculation**: Get precise costs for any date range
- **User-Specific Views**: Retrieve subscriptions by user
- **Admin Dashboard**: Special endpoints for administrative oversight
- **Soft Deletion**: Preserve data while marking subscriptions as deleted
- **REST API**: Standard HTTP endpoints for easy integration

- **Considerations**:
1. Subcriptions are not usually updated hence the absence of Update endpoint
2. The delete makes use of Patch method as it's a soft delete 


- **Limitations**:
1. Currently uses localhost:8080 as the base URL (needs configuration for production)
2. Currently lacks pagination //to be updated
3. Date formats are strict (MM-YYYY for subscription dates)

## API Endpoints

| Method | Endpoint                     | Description                          | Auth Required |
|--------|------------------------------|--------------------------------------|---------------|
| POST   | `/subscriptions`             | Create new subscription              | No            |
| GET    | `/subscriptions/user/{id}`   | Get user's subscriptions             | No            |
| GET    | `/subscriptions/{id}`        | Get specific subscription            | No            |
| POST   | `/subscriptions/{id}`        | Renew or extend a subscription       | No            |
| PATCH  | `/subscriptions/{id}`        | Soft-delete subscription             | No            |
| GET    | `/subscriptions`             | Get all subscriptions (admin only)   | Admin Key     |
| GET    | `/costs/{user_id}`           | Calculate subscription cost          | No            |

## Prerequisites

- Go 1.20+
- PostgreSQL (recommended)
- Docker
- Linux/macOS (or WSL on Windows)

## Installation

```bash
git clone https://github.com/Joshdike/subscriptions_aggregator.git
cd subscriptions_aggregator
go mod download