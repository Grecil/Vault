// REST API utility functions for FileVault

// Types
export interface FileInfo {
  id: string
  name: string
  size: number
  contentType: string
  uploadDate: string
  isPublic: boolean
  downloadCount?: number
  sha256?: string
}

export interface UserFilesResponse {
  files: FileInfo[]
  totalCount: number
  hasMore: boolean
}

export interface PrepareUploadResponse {
  uploadId: string
  presignedUrl?: string
  isDuplicate: boolean
  existingFile?: FileInfo
  expiresIn: number
}

export interface CompleteUploadResponse {
  success: boolean
  file: FileInfo
}

export interface DeleteFileResponse {
  success: boolean
  message: string
}

export interface DownloadUrlResponse {
  url: string
  expiresIn: number
}

export interface UpdateFileVisibilityResponse {
  success: boolean
  file: {
    id: string
    isPublic: boolean
  }
  shareLink?: string
}

// Batch upload types
export interface BatchFileRequest {
  filename: string
  size: number
  mime_type: string
  file_hash: string
}

export interface BatchFileResponse {
  file_hash: string
  status: 'upload_required' | 'duplicate' | 'quota_exceeded' | 'error'
  upload_id?: string
  presigned_url?: string
  existing_file?: FileInfo
  error?: string
}

export interface BatchPrepareResponse {
  batch_id: string
  files: BatchFileResponse[]
  quota_check: {
    total_size_required: number
    quota_available: boolean
    quota_exceeded?: number
  }
}

export interface BatchCompletedUpload {
  upload_id: string
  file_hash: string
  filename: string
  mime_type: string
}

export interface BatchCompleteResponse {
  batch_id: string
  completed_files: FileInfo[]
  errors?: string[]
}

// API configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

// Helper function to get auth headers
const getAuthHeaders = async (getToken: () => Promise<string | null>) => {
  const token = await getToken()
  return {
    'Content-Type': 'application/json',
    'Authorization': token ? `Bearer ${token}` : '',
  }
}

// API functions
export const apiClient = {
  // Get user files
  getUserFiles: async (
    getToken: () => Promise<string | null>,
    limit: number = 50,
    page: number = 1
  ): Promise<UserFilesResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(
      `${API_BASE_URL}/files?limit=${limit}&page=${page}`,
      {
        method: 'GET',
        headers,
      }
    )

    if (!response.ok) {
      throw new Error(`Failed to get files: ${response.statusText}`)
    }

    const data = await response.json()
    return {
      files: (data.files || []).map((file: any) => ({
        id: file.id,
        name: file.filename,
        size: file.size,
        contentType: file.mime_type,
        uploadDate: file.uploaded_at,
        isPublic: file.is_public,
        downloadCount: file.download_count || 0,
        sha256: file.file_hash
      })),
      totalCount: data.total || 0,
      hasMore: data.has_more || false
    }
  },

  // Prepare file upload
  prepareUpload: async (
    getToken: () => Promise<string | null>,
    filename: string,
    size: number,
    contentType: string,
    sha256: string
  ): Promise<PrepareUploadResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/upload-url`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        filename,
        size,
        mime_type: contentType,
        file_hash: sha256,
      }),
    })

    if (!response.ok) {
      throw new Error(`Failed to prepare upload: ${response.statusText}`)
    }

    const data = await response.json()
    return {
      uploadId: data.object_key || data.upload_id,
      presignedUrl: data.upload_url || data.presigned_url,
      isDuplicate: data.is_duplicate || false,
      existingFile: data.existing_file,
      expiresIn: data.expires_in || 3600
    }
  },

  // Complete file upload
  completeUpload: async (
    getToken: () => Promise<string | null>,
    objectKey: string,
    filename: string,
    mimeType: string,
    sha256: string
  ): Promise<CompleteUploadResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/complete`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        object_key: objectKey,
        filename,
        mime_type: mimeType,
        file_hash: sha256,
      }),
    })

    if (!response.ok) {
      throw new Error(`Failed to complete upload: ${response.statusText}`)
    }

    const data = await response.json()
    return {
      success: true,
      file: {
        id: data.id,
        name: data.filename,
        size: data.size,
        contentType: data.mime_type,
        uploadDate: data.uploaded_at,
        isPublic: data.is_public,
        sha256: data.file_hash
      }
    }
  },

  // Delete file
  deleteFile: async (
    getToken: () => Promise<string | null>,
    fileId: string
  ): Promise<DeleteFileResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/${fileId}`, {
      method: 'DELETE',
      headers,
    })

    if (!response.ok) {
      throw new Error(`Failed to delete file: ${response.statusText}`)
    }

    return response.json()
  },

  // Update file visibility
  updateFileVisibility: async (
    getToken: () => Promise<string | null>,
    fileId: string
  ): Promise<UpdateFileVisibilityResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/${fileId}/public`, {
      method: 'PATCH',
      headers,
    })

    if (!response.ok) {
      throw new Error(`Failed to update file visibility: ${response.statusText}`)
    }

    const data = await response.json()
    return {
      success: true,
      file: {
        id: fileId,
        isPublic: data.is_public
      },
      shareLink: data.share_link
    }
  },

  // Get download URL
  getDownloadUrl: async (
    getToken: () => Promise<string | null>,
    fileId: string
  ): Promise<DownloadUrlResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/${fileId}/download`, {
      method: 'GET',
      headers,
    })

    if (!response.ok) {
      throw new Error(`Failed to get download URL: ${response.statusText}`)
    }

    const data = await response.json()
    return {
      url: data.download_url,
      expiresIn: data.expires_in || 3600
    }
  },

  // Batch upload preparation
  batchPrepareUpload: async (
    getToken: () => Promise<string | null>,
    files: BatchFileRequest[]
  ): Promise<BatchPrepareResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/batch/prepare`, {
      method: 'POST',
      headers,
      body: JSON.stringify({ files })
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.error || `Batch prepare failed: ${response.statusText}`)
    }

    return response.json()
  },

  // Batch upload completion
  batchCompleteUpload: async (
    getToken: () => Promise<string | null>,
    batchId: string,
    completedUploads: BatchCompletedUpload[]
  ): Promise<BatchCompleteResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/batch/complete`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        batch_id: batchId,
        completed_uploads: completedUploads
      })
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.error || `Batch complete failed: ${response.statusText}`)
    }

    return response.json()
  },

  // Get share link for a public file
  getShareLink: async (
    getToken: () => Promise<string | null>,
    fileId: string
  ): Promise<{ share_link: string }> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/${fileId}/share-link`, {
      method: 'GET',
      headers,
    })

    if (!response.ok) {
      throw new Error(`Failed to get share link: ${response.statusText}`)
    }

    return response.json()
  },

  // Get user profile and storage info
  getUserProfile: async (
    getToken: () => Promise<string | null>
  ) => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/user/profile`, {
      method: 'GET',
      headers,
    })

    if (!response.ok) {
      throw new Error(`Failed to get user profile: ${response.statusText}`)
    }

    return response.json()
  },

  getUserStorage: async (
    getToken: () => Promise<string | null>
  ) => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/user/storage`, {
      method: 'GET',
      headers,
    })

    if (!response.ok) {
      throw new Error(`Failed to get storage info: ${response.statusText}`)
    }

    return response.json()
  },
}
