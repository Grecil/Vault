# FileVault - System Design

## High-Level Architecture

```
Frontend (React) → API Gateway (Kong) → FileVault API (Go) → Storage & Database
```

**Key Components:**
- **Frontend**: React app with Clerk authentication
- **API Gateway**: Kong handles auth, rate limiting, load balancing
- **Backend**: Single Go service with GraphQL API
- **Storage**: MinIO for files, PostgreSQL for metadata, Redis for caching

## Tech Stack

### Frontend
- **React 18 + TypeScript** - Modern, well-supported
- **Vite** - Fast development and builds
- **Clerk** - Handles all authentication (no custom auth code!)
- **Apollo Client** - GraphQL with caching
- **Shadcn/ui** - Beautiful, copy-paste components

### Backend
- **Go + Gin** - Fast HTTP server
- **GraphQL (gqlgen)** - Type-safe API with code generation
- **GORM** - Database ORM with auto-migrations
- **Clerk Go SDK** - Validates user tokens

### Infrastructure
- **Kong** - API gateway (handles auth, rate limiting)
- **PostgreSQL** - Metadata and user data
- **MinIO** - File storage (S3-compatible)
- **Redis** - Caching and rate limiting
- **Docker Compose** - Local development

## Database Design

### Core Tables
```sql
Users
- clerk_user_id (string, PK)
- role (enum: user, admin)
- storage_quota (bigint)
- storage_used (bigint)

FileHashes
- hash (string, PK) -- SHA256
- size (bigint)
- mime_type (string)
- reference_count (int)
- minio_key (string)

UserFiles
- id (UUID, PK)
- user_id (string, FK)
- file_hash (string, FK)
- filename (string)
- is_public (boolean)
- download_count (int)
- uploaded_at (timestamp)
```

## How It All Works

### File Upload Flow
1. User uploads file in React app
2. Kong validates JWT, forwards to backend
3. Backend calculates file hash
4. Check if hash exists in database
5. If exists: create user file record, increment reference count
6. If new: upload to MinIO, create hash + user file records
7. Return file metadata to frontend

### File Access Flow
1. User requests file
2. Backend checks ownership/public status
3. Generate MinIO presigned URL (temporary, secure link)
4. Return URL to frontend
5. User downloads directly from MinIO

### Admin Operations
1. Clerk roles determine admin access
2. Admin-specific GraphQL resolvers
3. Can view/manage all files and users
4. Storage analytics and system stats