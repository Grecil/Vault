# FileVault Local Setup

Simple setup guide for running FileVault locally using Docker Compose.

## Prerequisites

- Docker and Docker Compose installed
- Git installed
- Clerk account (free at [clerk.com](https://clerk.com))

## Quick Start

1. **Clone the repository:**
   ```bash
   git clone <your-repo-url>
   cd FileVault
   ```

2. **Set up environment variables:**
   ```bash
   cp Backend/env.example Backend/.env
   cp Frontend/env.example Frontend/.env
   ```
   - Make approapriate changes to both .env files

3. **Configure Clerk authentication:**
   - Go to [clerk.com](https://clerk.com) and create a new application
   - Copy your keys and update `Backend/.env`:
     ```
     CLERK_SECRET_KEY=sk_test_your_secret_key
     CLERK_PUBLISHABLE_KEY=pk_test_your_publishable_key
     ```

4. **Start all services:**
   ```bash
   docker-compose up -d
   ```

5. **Wait for services to initialize** (about 30 seconds)

6. **Configure MinIO bucket policy:**
   ```bash
   # Install MinIO client (if not already installed)
   # For Windows: Download from https://dl.min.io/client/mc/release/windows-amd64/mc.exe
   # For macOS: brew install minio/stable/mc
   # For Linux: wget https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x mc
   
   # Set up MinIO alias and bucket policy
   mc alias set myminio http://localhost:9000 minioadmin password123
   mc anonymous set-json bucket-policy.json myminio/files
   ```

7. **Open the application:**
   - Frontend: http://localhost:3000
   - API Swagger Documentation: http://localhost:8080/swagger/index.html
   - MinIO Console: http://localhost:9001 (admin/password123)

## Services

| Service | URL | Description |
|---------|-----|-------------|
| Frontend | http://localhost:3000 | React application |
| Backend API | http://localhost:8080 | Go API server |
| Swagger UI | http://localhost:8080/swagger/index.html | Interactive API documentation |
| PostgreSQL | localhost:5432 | Database |
| MinIO | http://localhost:9000 | File storage |
| MinIO Console | http://localhost:9001 | Storage management UI |

## API Documentation

Visit **http://localhost:8080/swagger/index.html** for complete API specifications including:
- Interactive endpoint testing
- Request/response examples
- Authentication requirements
- Error codes and descriptions

## Configuration

Default configuration works out of the box. To customize:

1. **Database settings** - Edit `docker-compose.yml` PostgreSQL section
2. **Storage settings** - Edit `Backend/.env` MinIO configuration
3. **Frontend settings** - Edit `Frontend/.env`

## Stopping Services

```bash
docker-compose down
```

To remove all data:
```bash
docker-compose down -v
```

## Troubleshooting

**Services not starting?**
- Check Docker is running
- Ensure ports 3000, 8080, 5432, 9000, 9001 are available

**Frontend can't connect to API?**
- Wait for all services to fully start (check `docker-compose logs`)
- Verify backend is running at http://localhost:8080/health

**File uploads not working?**
- Check MinIO console at http://localhost:9001
- Verify `files` bucket exists and is accessible
- Ensure bucket policy is applied: `mc anonymous set-json bucket-policy.json myminio/files`

**Need to reset everything?**
```bash
docker-compose down -v
docker-compose up -d
```

## Development

**Backend development:**
```bash
cd Backend
go mod download
go run cmd/server/main.go
```

**Frontend development:**
```bash
cd Frontend
bun install
bun run dev
```
