# FileVault Railway Deployment Guide

Simple step-by-step guide to deploy FileVault on Railway.

## Prerequisites

1. **Railway Account** - Sign up at [railway.app](https://railway.app)
2. **GitHub Repository** - Push your FileVault code to GitHub
3. **Clerk Account** - For authentication ([clerk.com](https://clerk.com))

## Architecture

```
Railway Services:
├── Frontend (React + Nginx)
├── Backend (Go API)  
├── PostgreSQL (Database)
└── MinIO (File Storage)
```

## Deployment Steps

### 1. Create Railway Project

1. Go to [railway.app](https://railway.app) and create a new project
2. Connect your GitHub repository
3. Deploy services in this order:

### 2. Deploy PostgreSQL

1. Click **"+ New"** → **"Database"** → **"PostgreSQL"**
2. Wait for deployment to complete

### 3. Deploy MinIO

1. Click **"+ New"** → **"Empty Service"**
2. Name it `minio`
3. Go to **Settings** → **Variables** and add:
   ```
   MINIO_ROOT_USER=minioadmin
   MINIO_ROOT_PASSWORD=your_secure_password_123
   ```
4. Go to **Settings** → **Deploy** and set:
   - **Image**: `minio/minio:latest`
   - **Start Command**: `server /data --console-address ':9001'`
5. Go to **Settings** → **Networking** and enable public networking
6. Deploy the service

### 4. Deploy Backend

1. Click **"+ New"** → **"GitHub Repo"** → Select your repository
2. Set **Root Directory**: `/Backend`
3. Go to **Settings** → **Variables** and add:
   ```
   DATABASE_URL=${{Postgres.DATABASE_URL}}
   PORT=8080
   GIN_MODE=release
   CLERK_SECRET_KEY=sk_live_your_clerk_secret_key
   MINIO_ENDPOINT=${{minio.RAILWAY_PRIVATE_DOMAIN}}:9000
   MINIO_ACCESS_KEY=${{minio.MINIO_ROOT_USER}}
   MINIO_SECRET_KEY=${{minio.MINIO_ROOT_PASSWORD}}
   MINIO_BUCKET=files
   MINIO_USE_SSL=false
   ```
4. Go to **Settings** → **Networking** and enable public networking
5. Deploy the service

### 5. Deploy Frontend

1. Click **"+ New"** → **"GitHub Repo"** → Select your repository  
2. Set **Root Directory**: `/Frontend`
3. Go to **Settings** → **Variables** and add:
   ```
   VITE_API_URL=https://${{backend.RAILWAY_PUBLIC_DOMAIN}}/api/v1
   VITE_CLERK_PUBLISHABLE_KEY=pk_live_your_clerk_publishable_key
   ```
4. Go to **Settings** → **Networking** and enable public networking
5. Deploy the service

### 6. Configure MinIO Bucket

After all services are running:

1. Go to your MinIO service URL (from Railway dashboard)
2. Login with your MINIO_ROOT_USER and MINIO_ROOT_PASSWORD
3. Create a bucket named `files`
4. Set bucket policy to public read for tagged objects

## Environment Variables Reference

### Backend Service
```bash
DATABASE_URL=${{Postgres.DATABASE_URL}}
PORT=8080
GIN_MODE=release
CLERK_SECRET_KEY=sk_live_your_secret_key
MINIO_ENDPOINT=${{minio.RAILWAY_PRIVATE_DOMAIN}}:9000
MINIO_ACCESS_KEY=${{minio.MINIO_ROOT_USER}}
MINIO_SECRET_KEY=${{minio.MINIO_ROOT_PASSWORD}}
MINIO_BUCKET=files
MINIO_USE_SSL=false
```

### Frontend Service
```bash
VITE_API_URL=https://${{backend.RAILWAY_PUBLIC_DOMAIN}}/api/v1
VITE_CLERK_PUBLISHABLE_KEY=pk_live_your_publishable_key
```

### MinIO Service
```bash
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your_secure_password_123
```

## Testing Your Deployment

1. **Backend Health**: Visit `https://your-backend.up.railway.app/health`
2. **Frontend**: Visit your frontend URL from Railway dashboard
3. **Upload Test**: Try uploading a file through the frontend

## Troubleshooting

### Common Issues

**Database Connection Errors:**
- Verify `DATABASE_URL` is correctly set
- Check PostgreSQL service is running

**MinIO Connection Errors:**
- Ensure MinIO service is running and accessible
- Verify `MINIO_ENDPOINT` includes port `:9000`
- Check MinIO credentials are correct

**CORS Errors:**
- Verify backend URL is correct in frontend env vars
- Check both services have public networking enabled

**File Upload Errors:**
- Ensure MinIO bucket `files` exists
- Verify MinIO is accessible from backend service

### Debug Commands

View service logs in Railway dashboard:
- Click on service → **Deployments** → **View Logs**

## Cost Estimation

**Monthly costs (approximate):**
- Railway Hobby Plan: $5/month per service
- PostgreSQL: $5/month  
- **Total: ~$25/month for 4 services**

## Security Notes

1. **Never commit secrets** - Use Railway environment variables
2. **Use strong passwords** for MinIO and database
3. **Keep services private** except frontend/backend public networking
4. **Rotate Clerk keys** regularly for production

---

**Need help?** Check Railway documentation or create an issue in the repository.