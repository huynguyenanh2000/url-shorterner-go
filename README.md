# URL Shortener API

A high-performance URL shortening service containerized with Docker, featuring PostgreSQL for persistent storage and Redis for high-speed caching.

## 🚀 Deployment Guide

Follow these steps to deploy the application environment.

### 1. Environment Configuration
Create a `.env` file in the root directory and populate it with your configuration details:

```env
# Database Credentials
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_ROOT_PASSWORD=

# Connection Pooling
DB_MAX_OPEN_CONNS=
DB_MAX_IDLE_CONNS=
DB_MAX_IDLE_TIME=

# Redis Configuration
REDIS_PW=
REDIS_DB=

# Environment Settings
ENV= ```

## 🚀 Execution & Deployment

### 2. Launch Services
Once your `.env` file is configured, use **Docker Compose** to pull the necessary images and start the infrastructure in the background.

```bash
# Build and start all containers in detached mode
docker compose up -d

### 3. Finalize API Setup

To ensure the application correctly initializes its connection pools and applies environment variables from the `.env` file, restart the API container:

```bash
docker restart url-shortener-api
