# FileVault Railway Deployment Guide

## Prerequisites

1. **Railway Account** - Sign up at [railway.app](https://railway.app)
2. **GitHub Repository** - Your FileVault code should be in a GitHub repository
3. **Clerk Account** - For authentication

## Architecture Overview

```
Railway Services:
├── Frontend (React + Nginx)
├── Backend (Go API)  
├── PostgreSQL (Database)
└── MinIO (File Storage)
```

## Step 1: Prepare Your Repository

1. **Add Dockerfiles** (already created):
   - `Backend/Dockerfile.railway`
   - `Frontend/Dockerfile.railway`
   - `Frontend/nginx.railway.conf`

2. **Update Backend Dependencies**:
   ```bash
   cd Backend
   go mod tidy
   ```

## Step 2: Deploy to Railway

### Option A: Using Railway CLI (Recommended)

1. **Install Railway CLI**:
   ```bash
   npm install -g @railway/cli
   ```

2. **Login and Initialize**:
   ```bash
   railway login
   railway init
   ```

3. **Deploy Services**:

   **Backend Service:**
   ```bash
   # Create backend service
   railway service create backend
   railway service use backend
   
   # Set environment variables
   railway variables set GIN_MODE=release
   railway variables set CLERK_SECRET_KEY="your_clerk_secret_key"
   railway variables set MINIO_BUCKET="files"
   railway variables set MINIO_USE_SSL="false"
   
   # Deploy
   railway up --service backend --dockerfile Backend/Dockerfile.railway
   ```

   **Frontend Service:**
   ```bash
   # Create frontend service
   railway service create frontend
   railway service use frontend
   
   # Set environment variables (replace with your domains)
   railway variables set VITE_CLERK_PUBLISHABLE_KEY="your_clerk_publishable_key"
   
   # Deploy
   railway up --service frontend --dockerfile Frontend/Dockerfile.railway
   ```

   **PostgreSQL:**
   ```bash
   railway add postgresql
   ```

   **MinIO:**
   ```bash
   # Create MinIO service
   railway service create minio
   railway service use minio
   
   # Set environment variables
   railway variables set MINIO_ROOT_USER="minioadmin"
   railway variables set MINIO_ROOT_PASSWORD="your_secure_password_here"
   
   # Deploy MinIO
   railway up --service minio --image minio/minio:latest
   ```


### Option B: Using Railway Dashboard

1. **Go to [railway.app](https://railway.app)** and create a new project
2. **Connect your GitHub repository**
3. **Add services one by one**:
   - Backend (set root directory to `/Backend`)
   - Frontend (set root directory to `/Frontend`)
   - PostgreSQL (from template)
   - MinIO (from Docker image)
  
FileVault uses object tags to control file access:
- **Private files**: No tags → require authentication
- **Public files**: Tagged with `public=true` → accessible without authentication

## 1. Quick Setup

Create `bucket-policy.json`:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::files/*",
      "Condition": {
        "StringEquals": {
          "s3:ExistingObjectTag/public": "true"
        }
      }
    }
  ]
}
```

Apply the policy:
```bash
# Setup mc client
mc alias set myminio http://localhost:9000 minioadmin minioadmin123

# Apply bucket policy
mc anonymous set-json bucket-policy.json myminio/files
```

## 2. How It Works

**Public Files**: `http://localhost:9000/files/{hash}` (clean URLs)  
**Private Files**: Presigned URLs with 1-minute TTL

When you toggle a file:
- **Make Public**: Backend sets `public=true` tag → public URL works
- **Make Private**: Backend removes tag → public URL returns "Access Denied"

## 3. Verify Setup

Check policy is active:
```bash
mc anonymous get myminio/files
# Should show: Access permission for `myminio/files` is `custom`
```

Test with a public file:
```bash
# Check object tags
mc tag list myminio/files/{hash}
# Should show: public=true

# Test direct access
curl http://localhost:9000/files/{hash}
# Should download the file
```

## 4. Troubleshooting

**Access still denied for public files?**
- Verify bucket policy applied: `mc anonymous get myminio/files` shows `custom`
- Check object has tag: `mc tag list myminio/files/{hash}` shows `public=true`
- Restart MinIO if needed

## Step 3: Configure Environment Variables

### Backend Service Variables:
```bash
# Database (auto-provided by Railway)
DATABASE_URL=${Postgres.DATABASE_URL}

# Server
PORT=8080
GIN_MODE=release

# Authentication
CLERK_SECRET_KEY=sk_live_your_secret_key

# MinIO (Reference other services)
MINIO_ENDPOINT=${minio.RAILWAY_PRIVATE_DOMAIN}:9000
MINIO_ACCESS_KEY=${minio.MINIO_ROOT_USER}
MINIO_SECRET_KEY=${minio.MINIO_ROOT_PASSWORD}
MINIO_BUCKET=filevault-files
MINIO_USE_SSL=false

```

### Frontend Service Variables:
```bash
# API URL (replace with your backend domain)
VITE_API_URL=https://your-backend.up.railway.app

# MinIO URL (replace with your MinIO domain)  
VITE_MINIO_URL=https://your-minio.up.railway.app

# Authentication
VITE_CLERK_PUBLISHABLE_KEY=pk_live_your_publishable_key
```

### MinIO Service Variables:
```bash
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your_secure_password_here
```

## Step 4: Configure Networking

1. **Backend Service**:
   - Enable public networking
   - Note the domain (e.g., `your-backend.up.railway.app`)

2. **Frontend Service**:
   - Enable public networking
   - Note the domain (e.g., `your-frontend.up.railway.app`)

3. **MinIO Service**:
   - Enable public networking for presigned URLs
   - Note the domain (e.g., `your-minio.up.railway.app`)

4. **PostgreSQL**:
   - Keep private networking only

## Step 5: Update Frontend Configuration

Update your frontend to use the Railway domains:

```typescript
// Frontend/src/lib/api.ts
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const MINIO_BASE_URL = import.meta.env.VITE_MINIO_URL || 'http://localhost:9000';
```

## Step 6: Custom Domains (Optional)

1. **In Railway Dashboard**:
   - Go to each service
   - Click "Settings" → "Domains"
   - Add your custom domain
   - Update DNS records as instructed

2. **Update Environment Variables**:
   - Update `VITE_API_URL` and `VITE_MINIO_URL` to use custom domains

## Step 7: Testing

1. **Check Service Health**:
   - Backend: `https://your-backend.up.railway.app/health`
   - Frontend: `https://your-frontend.up.railway.app`
   - MinIO: `https://your-minio.up.railway.app/minio/health/live`

2. **Test File Upload**:
   - Go to your frontend
   - Try uploading a file
   - Check MinIO and database for the file

## Step 8: Monitoring & Logs

1. **Railway Dashboard**:
   - View logs for each service
   - Monitor resource usage
   - Set up alerts

2. **Health Checks**:
   - Railway automatically monitors service health
   - Failed deployments are automatically rolled back

## Cost Optimization

1. **Resource Limits**:
   - Set appropriate CPU/memory limits for each service
   - Scale down non-critical services

2. **Auto-Sleep**:
   - Enable auto-sleep for development environments
   - Keep production services always running

## Troubleshooting

### Common Issues:

1. **Database Connection Errors**:
   - Check `DATABASE_URL` format
   - Ensure SSL mode is correct

2. **MinIO Connection Errors**:
   - Verify `MINIO_ENDPOINT` includes port (`:9000`)
   - Check if `MINIO_USE_SSL` is set correctly

3. **CORS Errors**:
   - Ensure frontend and backend domains are correctly configured
   - Check Clerk CORS settings

4. **File Upload Errors**:
   - Verify MinIO bucket is created
   - Check MinIO credentials
   - Ensure presigned URLs are accessible

### Debug Commands:

```bash
# View logs
railway logs --service backend
railway logs --service frontend
railway logs --service minio

# Connect to database
railway connect postgresql

# Check environment variables
railway variables --service backend
```

## Scaling

Railway automatically scales based on:
- CPU usage
- Memory usage
- Request volume

You can also manually scale:
```bash
railway scale --service backend --replicas 2
```

## Backup Strategy

1. **Database Backups**:
   - Railway automatically backs up PostgreSQL
   - Can restore from dashboard

2. **File Backups**:
   - MinIO data is persistent on Railway volumes
   - Consider external backup for critical files

## Security Considerations

1. **Environment Variables**:
   - Never commit secrets to repository
   - Use Railway's environment variable management

2. **Network Security**:
   - Keep database and Redis private
   - Only expose necessary services publicly

3. **Authentication**:
   - Use Clerk for user authentication
   - Validate JWT tokens on backend

4. **File Access**:
   - Use presigned URLs for file access
   - Implement proper authorization checks

## Cost Estimation

**Monthly Costs (approximate):**
- Railway Pro Plan: $5/month
- PostgreSQL: $5/month
- MinIO: $5/month (depending on storage)
- **Total: ~$15/month**
