// REST API utility functions for FileVault

// Types
export interface FileInfo {
  id: string
  name: string
  size: number
  contentType: string
  uploadDate: string
  isPublic: boolean
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
}

// API configuration
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

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
    offset: number = 0
  ): Promise<UserFilesResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(
      `${API_BASE_URL}/files?limit=${limit}&offset=${offset}`,
      {
        method: 'GET',
        headers,
      }
    )

    if (!response.ok) {
      throw new Error(`Failed to get files: ${response.statusText}`)
    }

    return response.json()
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
    const response = await fetch(`${API_BASE_URL}/files/prepare-upload`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        filename,
        size,
        contentType,
        sha256,
      }),
    })

    if (!response.ok) {
      throw new Error(`Failed to prepare upload: ${response.statusText}`)
    }

    return response.json()
  },

  // Complete file upload
  completeUpload: async (
    getToken: () => Promise<string | null>,
    uploadId: string,
    sha256: string
  ): Promise<CompleteUploadResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/complete-upload`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        uploadId,
        sha256,
      }),
    })

    if (!response.ok) {
      throw new Error(`Failed to complete upload: ${response.statusText}`)
    }

    return response.json()
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
    fileId: string,
    isPublic: boolean
  ): Promise<UpdateFileVisibilityResponse> => {
    const headers = await getAuthHeaders(getToken)
    const response = await fetch(`${API_BASE_URL}/files/${fileId}/visibility`, {
      method: 'PATCH',
      headers,
      body: JSON.stringify({
        isPublic,
      }),
    })

    if (!response.ok) {
      throw new Error(`Failed to update file visibility: ${response.statusText}`)
    }

    return response.json()
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

    return response.json()
  },
}
