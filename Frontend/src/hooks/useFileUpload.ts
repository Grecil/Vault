import { useState, useCallback } from 'react'
import { useUser } from '@clerk/clerk-react'
import { calculateSHA256, formatFileSize } from '../utils/crypto'
import { apiClient } from '../lib/api'

export interface UploadingFile {
  file: File
  id: string
  progress: number
  status: 'pending' | 'hashing' | 'preparing' | 'uploading' | 'completed' | 'failed' | 'duplicate'
  error: string | null
  sha256: string
  existingFile?: any
  fileInfo?: any
}

export const useFileUpload = (maxFileSize: number = 10 * 1024 * 1024) => {
  const [uploadingFiles, setUploadingFiles] = useState<UploadingFile[]>([])
  const { getToken } = useUser()

  const updateFileStatus = useCallback((fileId: string, updates: Partial<UploadingFile>) => {
    setUploadingFiles(prev => 
      prev.map(file => 
        file.id === fileId ? { ...file, ...updates } : file
      )
    )
  }, [])

  const uploadFileToMinio = async (url: string, file: File, onProgress: (progress: number) => void) => {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      xhr.open('PUT', url)
      xhr.setRequestHeader('Content-Type', file.type || 'application/octet-stream')

      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          const percent = (event.loaded / event.total) * 100
          onProgress(percent)
        }
      }

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve(xhr.response)
        } else {
          reject(new Error(`Upload failed with status ${xhr.status}: ${xhr.statusText}`))
        }
      }

      xhr.onerror = () => {
        reject(new Error('Network error during upload'))
      }

      xhr.send(file)
    })
  }

  const processFile = async (file: File) => {
    const fileId = Math.random().toString(36).substring(7)
    const fileEntry: UploadingFile = {
      file,
      id: fileId,
      progress: 0,
      status: 'pending',
      error: null,
      sha256: '',
    }

    setUploadingFiles(prev => [...prev, fileEntry])
    updateFileStatus(fileId, { status: 'hashing' })

    try {
      // Validate file size
      if (file.size > maxFileSize) {
        updateFileStatus(fileId, { 
          status: 'failed', 
          error: `File size exceeds ${formatFileSize(maxFileSize)} limit.` 
        })
        return
      }

      // Calculate SHA256
      const sha256 = await calculateSHA256(file)
      updateFileStatus(fileId, { sha256, status: 'preparing' })

      // Prepare upload
      const data = await apiClient.prepareUpload(
        getToken,
        file.name,
        file.size,
        file.type || 'application/octet-stream',
        sha256
      )

      const { uploadId, presignedUrl, isDuplicate, existingFile } = data

      if (isDuplicate && existingFile) {
        updateFileStatus(fileId, { status: 'duplicate', existingFile })
        console.log('File is duplicate:', existingFile)
        return
      }

      if (!presignedUrl) {
        throw new Error("Failed to get presigned URL.")
      }

      updateFileStatus(fileId, { status: 'uploading' })

      // Upload to MinIO
      await uploadFileToMinio(presignedUrl, file, (progress) => {
        updateFileStatus(fileId, { progress })
      })

      // Complete upload
      const completeData = await apiClient.completeUpload(getToken, uploadId, sha256)

      if (completeData.success) {
        updateFileStatus(fileId, { 
          status: 'completed', 
          fileInfo: completeData.file 
        })
        console.log('Upload completed:', completeData.file)
        return completeData.file
      } else {
        throw new Error("Failed to complete upload on backend.")
      }

    } catch (error: any) {
      console.error("Upload error:", error)
      updateFileStatus(fileId, { 
        status: 'failed', 
        error: error.message || 'Upload failed' 
      })
    }
  }

  const removeFile = useCallback((fileId: string) => {
    setUploadingFiles(prev => prev.filter(file => file.id !== fileId))
  }, [])

  const clearCompleted = useCallback(() => {
    setUploadingFiles(prev => prev.filter(file => file.status !== 'completed'))
  }, [])

  const clearAll = useCallback(() => {
    setUploadingFiles([])
  }, [])

  return {
    uploadingFiles,
    processFile,
    removeFile,
    clearCompleted,
    clearAll,
    isUploading: uploadingFiles.some(file => file.status === 'uploading'),
    hasErrors: uploadingFiles.some(file => file.status === 'failed'),
  }
}
