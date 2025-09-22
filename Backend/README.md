Backend for the FileVault app

# MinIO Bucket Policy Setup for Tag-Based Access Control

## Overview

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