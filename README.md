# <img src="Frontend\public\vault.svg"> Vault

A secure, modern file storage and sharing service built with Go and React. Upload, manage, and share files with ease. You can access it at [vault.grecil.in](https://vault.grecil.in/).

## Features

- **Secure Authentication** - Powered by Clerk
- **File Management** - Upload, download, delete files
- **File Sharing** - Make a file private or public
- **Search & Filtering** - Fuzzy Search and Sorting
- **Object Storage** - MinIO backend for scalability
- **Storage Quotas** - Per-user storage limits
- **Rate Limiting** - Token Bucket Algorithm
- **Modern UI** - Responsive React frontend

## Documentation

| Document | Description |
|----------|-------------|
| [Setup Guide](setup.md) | Complete local development setup with Docker |
| [Railway Deployment](railway_deployment.md) | Production deployment on Railway |
| [Design Document](design.md) | Architecture and technical specifications |
| [API Documentation](http://localhost:8080/swagger/index.html) | Interactive API specs (when running) |

## Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   React     │    │     Go      │    │ PostgreSQL  │
│  Frontend   │◄──►│   Backend   │◄──►│  Database   │
└─────────────┘    └─────────────┘    └─────────────┘
                           │
                           ▼
                   ┌─────────────┐
                   │    MinIO    │
                   │   Storage   │
                   └─────────────┘
```

## Tech Stack

**Frontend:**
- React 19 + TypeScript
- Tailwind CSS
- Vite
- Clerk Authentication

**Backend:**
- Go 1.25 + Gin framework
- PostgreSQL database
- MinIO object storage
- Swagger documentation