# CloudForge

CloudForge is a Docker infrastructure management REST API built with Go. It provides centralized management of Docker containers and images, user authentication, role-based access control, and audit logging.

## Features

- JWT-based user authentication.
- Role-Based Access Control (RBAC).
- User management.
- Docker container management: list, start, stop, and restart.
- Container logs and resource statistics.
- Docker image management: list, inspect, pull, and remove.
- Audit logging.
- API health check.
- Automatic database migrations.

## Tech Stack

- Go 1.27.4
- Gin
- PostgreSQL 17
- GORM
- Docker Engine API
- JWT

## Project Structure

```text
cloudforge/
├── backend/
│   ├── cmd/api/
│   ├── internal/
│   │   ├── database/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── models/
│   │   └── services/
│   ├── .env.example
│   ├── go.mod
│   └── go.sum
├── deployments/docker/docker-compose.yml
├── .gitignore
├── LICENSE
├── Makefile
└── README.md
```

## Prerequisites

- Go 1.27.4
- Docker Desktop or Docker Engine
- Git
- Make

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/VerryHdyt/cloudforge.git
cd cloudforge
```

### 2. Configure environment variables

```bash
cp backend/.env.example backend/.env
```

Edit `backend/.env`:

```env
APP_ENV=development
APP_PORT=8080
DATABASE_URL=host=localhost user=cloudforge password=YOUR_PASSWORD dbname=cloudforge port=5433 sslmode=disable TimeZone=Asia/Jakarta
JWT_SECRET=YOUR_RANDOM_SECRET
```

Use a strong, unique JWT secret. Never commit your `.env` file.

### 3. Start PostgreSQL

```bash
docker compose -f deployments/docker/docker-compose.yml up -d
```

PostgreSQL is exposed on host port `5433`.

### 4. Start the API

Ensure Docker Desktop or Docker Engine is running and accessible.

```bash
make run
```

The API runs at `http://localhost:8080`. Database migrations run automatically at startup.

## API Endpoints

All versioned endpoints use `/api/v1`.

| Method | Endpoint | Access |
|---|---|---|
| GET | `/health` | Public |
| GET | `/api/v1/info` | Public |
| POST | `/api/v1/auth/register` | Public |
| POST | `/api/v1/auth/login` | Public |
| POST | `/api/v1/users` | Admin |
| GET | `/api/v1/users` | Admin |
| GET | `/api/v1/users/:id` | Admin |
| PUT | `/api/v1/users/:id` | Admin |
| DELETE | `/api/v1/users/:id` | Admin |
| GET | `/api/v1/audit-logs` | Admin |
| GET | `/api/v1/containers` | Admin |
| POST | `/api/v1/containers/:id/start` | Admin |
| POST | `/api/v1/containers/:id/stop` | Admin |
| POST | `/api/v1/containers/:id/restart` | Admin |
| GET | `/api/v1/containers/:id/logs` | Admin |
| GET | `/api/v1/containers/:id/stats` | Admin |
| GET | `/api/v1/images` | Admin |
| GET | `/api/v1/images/:id` | Admin |
| POST | `/api/v1/images/pull` | Admin |
| DELETE | `/api/v1/images/:id` | Admin |

Admin endpoints require a valid JWT and the `admin` role.

## Development Commands

| Command | Description |
|---|---|
| `make help` | Show available commands |
| `make run` | Run the API |
| `make build` | Build Go packages |
| `make test` | Run unit tests |
| `make test-race` | Run tests with race detector |
| `make vet` | Run Go vet |
| `make vuln` | Run govulncheck |
| `make fmt` | Format Go files |
| `make tidy` | Clean Go dependencies |
| `make check` | Run all quality checks |

## Testing

Run all quality checks:

```bash
make check
```

Or run individual checks:

```bash
make test
make test-race
make vet
make vuln
```

## Security

- Never commit environment files, credentials, or secrets.
- Use strong credentials outside local development.
- Restrict access to the Docker Engine.
- Review development credentials before deploying to production.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).