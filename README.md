# CircleConnect Feed Service

This service provides personalized feed functionality for the CircleConnect platform.

## Features

- Personalized feed based on user's joined communities
- Recommendation system based on post popularity, user preferences, and demographics
- Feed sorting options (by date, relevance, popularity)
- User preference management

## API Endpoints

### Feed Endpoints

- `GET /api/feed` - Retrieve personalized feed for authenticated user
  - Query parameters:
    - `sort_by` - Sort method (date, relevance, popular)
    - `community_id` - Filter by community
    - `tags` - Filter by tags
    - `page` - Page number
    - `limit` - Items per page

- `GET /api/feed/recommended` - Get recommended posts
  - Query parameters:
    - `page` - Page number
    - `limit` - Items per page

### Preference Endpoints

- `GET /api/feed/preferences` - Get user feed preferences
- `PUT /api/feed/preferences` - Update user feed preferences

## Getting Started

### Prerequisites

- Go 1.21+
- MongoDB
- PostgreSQL

### Running Locally

1. Clone the repository
2. Copy `.env.example` to `.env` and modify as needed
3. Run `go mod download` to install dependencies
4. Run `go run main.go` to start the service

### Running with Docker

```bash
docker-compose up -d
```

## Environment Variables

- `PORT` - Server port (default: 4004)
- `MONGO_URI` - MongoDB connection URI
- `POSTGRES_URI` - PostgreSQL connection URI
- `MONGO_DB_NAME` - MongoDB database name
- `ENVIRONMENT` - Runtime environment (development/production)
- `JWT_SECRET` - Secret key for JWT validation
- `USER_SERVICE_URL` - User service endpoint URL
- `POST_SERVICE_URL` - Post service endpoint URL
- `COMMUNITY_SERVICE_URL` - Community service endpoint URL 