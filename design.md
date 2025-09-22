## High-Level Architecture

**Key Components:**
- **Frontend**: Modern React SPA with TypeScript and Tailwind CSS
- **Backend**: RESTful API service built with Go and Gin framework
- **Authentication**: Clerk handles all user authentication and session management
- **Storage**: MinIO for file storage, PostgreSQL for metadata
- **Deployment**: Railway platform with containerized services

## Tech Stack

### Frontend
- **React 19 + TypeScript** - Modern UI framework with type safety
- **Vite** - Fast development server and build tool
- **Tailwind CSS** - Utility-first CSS framework for styling
- **Clerk React** - Complete authentication solution
- **React Router** - Client-side routing
- **React Dropzone** - Drag & drop file upload interface
- **Fuse.js** - Fuzzy search for file filtering

### Backend
- **Go 1.25** - High-performance compiled language
- **Gin Framework** - Fast HTTP web framework
- **GORM** - Object-relational mapping with PostgreSQL
- **Clerk Go SDK** - Server-side authentication validation
- **MinIO Go SDK** - S3-compatible object storage client
- **UUID** - Unique identifier generation

### Infrastructure
- **PostgreSQL** - Primary database for metadata and user data
- **MinIO** - S3-compatible object storage for files
- **Railway** - Cloud deployment platform
- **Nginx** - Static file serving and reverse proxy
- **Docker** - Containerization for consistent deployments

## Database Design

### Core Tables
```sql
Users
- id (string, PK) -- Clerk user ID
- role (enum: user, admin)
- storage_quota (bigint)
- storage_used (bigint)
- created_at, updated_at, deleted_at (timestamps)

FileHashes  
- hash (string, PK) -- SHA256 file hash for deduplication
- size (bigint)
- mime_type (string)
- reference_count (int) -- How many users reference this file
- minio_key (string) -- Object key in MinIO storage
- created_at, updated_at (timestamps)

UserFiles
- id (UUID, PK)
- user_id (string, FK) -- References Users.id
- file_hash (string, FK) -- References FileHashes.hash
- filename (string) -- User-defined filename
- is_public (boolean) -- Public sharing flag
- download_count (int) -- Download statistics
- uploaded_at, updated_at, deleted_at (timestamps)

ShareLinks
- id (string, PK) -- Short random ID (8 chars)
- user_file_id (UUID, FK) -- References UserFiles.id
- created_at, updated_at, deleted_at (timestamps)
```

### Key Design Principles
- **File Deduplication**: Same file content stored once in MinIO, referenced by multiple users
- **Reference Counting**: Files are only deleted from storage when no users reference them
- **Public Sharing**: Files can be made public with clean shareable URLs
- **Storage Quotas**: Per-user storage limits with usage tracking

## How It All Works

### File Upload Flow
1. **Frontend**: User drags/drops files into React upload component
2. **Hash Calculation**: Client calculates SHA256 hash of file content
3. **Upload Request**: Frontend requests presigned upload URL from backend
4. **Authentication**: Clerk token validated, user quota checked
5. **Deduplication Check**: Backend checks if file hash already exists
6. **Storage**:
   - If new: Upload file to MinIO, create FileHash record
   - If exists: Skip upload, increment reference count
7. **Metadata**: Create UserFile record linking user to file
8. **Response**: Return file metadata to frontend

### File Access & Sharing Flow
1. **File List**: Backend returns user's files with metadata
2. **Download**: Generate MinIO presigned URL for secure direct download with short ttl
3. **Public Sharing**: 
   - User toggles file to public
   - Backend creates ShareLink with short random ID and sets `public=true` tag on MinIO object
   - Clean shareable URL: `/share/{id}`

### User Interface Features
- **Dashboard**: File management with grid/list views
- **Drag & Drop**: Intuitive file upload with progress indicators
- **Search & Filter**: Real-time file search with fuzzy matching
- **Dark Mode**: Theme toggle with system preference detection
- **Storage Stats**: Visual storage usage and quota information
- **Landing Page**: Marketing site with feature showcase

### Admin Operations
1. **Role-Based Access**: Clerk manages user roles (user/admin)
2. **Admin Panel**: Separate routes for user and file management
3. **User Management**: View all users, update roles and quotas
4. **System Statistics**: Storage analytics and platform metrics
5. **File Oversight**: View and manage all files across users

### Security & Performance
- **Authentication**: Clerk handles all auth flows and session management
- **Rate Limiting**: Built-in request rate limiting per user
- **File Deduplication**: Reduces storage costs by storing identical files once
- **Presigned URLs**: Direct MinIO access without backend bottlenecks
- **Reference Counting**: Files physically deleted only when no users reference them


## Deployment Architecture

### Railway Platform Services
```
Production Environment:
├── Frontend Service (React + Nginx)
│   ├── Static file serving
│   ├── SPA routing support
│   └── Environment: VITE_API_URL, VITE_CLERK_PUBLISHABLE_KEY
├── Backend Service (Go API)
│   ├── RESTful API endpoints
│   ├── Clerk authentication
│   ├── File upload/download logic
│   └── Environment: DATABASE_URL, CLERK_SECRET_KEY, MINIO_*
├── PostgreSQL Database
│   ├── Managed database service
│   ├── Automatic backups
│   └── Connection pooling
└── MinIO Storage Service
    ├── S3-compatible object storage
    └── Persistent volume mounting
```

### Service Communication
- **Frontend ↔ Backend**: HTTPS REST API calls
- **Backend ↔ Database**: Direct PostgreSQL connection
- **Backend ↔ MinIO**: Internal network communication
- **Frontend ↔ MinIO**: Direct presigned URL downloads
- **Authentication**: Clerk manages tokens and validation

### Environment Configuration
- **Development**: Local Docker Compose setup
- **Production**: Railway platform with environment variables
- **Security**: All sensitive data in environment variables
- **Scaling**: Horizontal scaling ready for backend services