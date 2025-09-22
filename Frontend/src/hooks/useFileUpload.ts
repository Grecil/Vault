import { useState, useCallback } from 'react'
import { useAuth } from '@clerk/clerk-react'
import { calculateSHA256, formatFileSize } from '../utils/crypto'
import { apiClient, type BatchFileRequest, type BatchCompletedUpload } from '../lib/api'

export interface UploadingFile {
  file: File
  id: string
  progress: number
  status: 'pending' | 'hashing' | 'preparing' | 'uploading' | 'completed' | 'failed' | 'duplicate' | 'quota_exceeded'
  error: string | null
  sha256: string
  uploadId?: string
  existingFile?: any
  fileInfo?: any
}

export const useFileUpload = (maxFileSize: number = 10 * 1024 * 1024) => {
  const [uploadingFiles, setUploadingFiles] = useState<UploadingFile[]>([])
  const [batchId, setBatchId] = useState<string | null>(null)
  const { getToken } = useAuth()

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

  // Process multiple files as a batch
  const processFiles = async (files: File[]) => {
    // Validate and add files to the uploading list
    const validFiles: File[] = []
    const fileEntries: UploadingFile[] = []

    for (const file of files) {
      const fileId = Math.random().toString(36).substring(7)
      const fileEntry: UploadingFile = {
        file,
        id: fileId,
        progress: 0,
        status: 'pending',
        error: null,
        sha256: '',
      }

      // Validate file size
      if (file.size > maxFileSize) {
        fileEntry.status = 'failed'
        fileEntry.error = `File size exceeds ${formatFileSize(maxFileSize)} limit.`
        fileEntries.push(fileEntry)
        continue
      }

      validFiles.push(file)
      fileEntries.push(fileEntry)
    }

    // Add all files to state at once
    setUploadingFiles(prev => [...prev, ...fileEntries])

    if (validFiles.length === 0) {
      return // No valid files to process
    }

    try {
      // Phase 1: Hash all files
      const batchRequests: BatchFileRequest[] = []
      
      for (let i = 0; i < validFiles.length; i++) {
        const file = validFiles[i]
        const fileEntry = fileEntries.find(entry => entry.file === file)
        if (!fileEntry) continue

        updateFileStatus(fileEntry.id, { status: 'hashing' })
        
        const sha256 = await calculateSHA256(file)
        // Update both state and local array
        updateFileStatus(fileEntry.id, { sha256, status: 'preparing' })
        fileEntry.sha256 = sha256 // Update local array too
        
        batchRequests.push({
          filename: file.name,
          size: file.size,
          mime_type: file.type || 'application/octet-stream',
          file_hash: sha256
        })
      }

      // Phase 2: Batch prepare upload
      const batchResponse = await apiClient.batchPrepareUpload(getToken, batchRequests)
      setBatchId(batchResponse.batch_id)

      // Phase 3: Update file statuses based on server response
      const filesToUpload: { fileEntry: UploadingFile; uploadId: string; presignedUrl: string }[] = []
      
      for (const responseFile of batchResponse.files) {
        const fileEntry = fileEntries.find(entry => entry.sha256 === responseFile.file_hash)
        if (!fileEntry) continue

        switch (responseFile.status) {
          case 'duplicate':
            updateFileStatus(fileEntry.id, { 
              status: 'duplicate', 
              existingFile: responseFile.existing_file 
            })
            break
          
          case 'quota_exceeded':
            updateFileStatus(fileEntry.id, { 
              status: 'quota_exceeded', 
              error: responseFile.error || 'Storage quota exceeded' 
            })
            break
          
          case 'upload_required':
            if (responseFile.upload_id && responseFile.presigned_url) {
              updateFileStatus(fileEntry.id, { 
                status: 'uploading',
                uploadId: responseFile.upload_id 
              })
              filesToUpload.push({
                fileEntry,
                uploadId: responseFile.upload_id,
                presignedUrl: responseFile.presigned_url
              })
            }
            break
          
          case 'error':
            updateFileStatus(fileEntry.id, { 
              status: 'failed', 
              error: responseFile.error || 'Unknown error' 
            })
            break
        }
      }

      // Phase 4: Upload files that need uploading
      const uploadPromises = filesToUpload.map(async ({ fileEntry, uploadId, presignedUrl }) => {
        try {
          await uploadFileToMinio(presignedUrl, fileEntry.file, (progress) => {
            updateFileStatus(fileEntry.id, { progress })
          })
          return { fileEntry, uploadId, success: true }
        } catch (error: any) {
          updateFileStatus(fileEntry.id, { 
            status: 'failed', 
            error: error.message || 'Upload failed' 
          })
          return { fileEntry, uploadId, success: false }
        }
      })

      const uploadResults = await Promise.all(uploadPromises)

      // Phase 5: Complete successful uploads
      const completedUploads: BatchCompletedUpload[] = uploadResults
        .filter(result => result.success)
        .map(result => ({
          upload_id: result.uploadId,
          file_hash: result.fileEntry.sha256,
          filename: result.fileEntry.file.name,
          mime_type: result.fileEntry.file.type || 'application/octet-stream'
        }))

      if (completedUploads.length > 0) {
        const completeResponse = await apiClient.batchCompleteUpload(
          getToken, 
          batchResponse.batch_id, 
          completedUploads
        )

        // Update completed files
        for (let i = 0; i < completedUploads.length; i++) {
          const completedUpload = completedUploads[i]
          const fileEntry = fileEntries.find(entry => entry.sha256 === completedUpload.file_hash)
          const completedFile = completeResponse.completed_files[i]
          
          if (fileEntry && completedFile) {
            updateFileStatus(fileEntry.id, { 
              status: 'completed', 
              fileInfo: completedFile 
            })
          }
        }
      }

    } catch (error: any) {
      console.error("Batch upload error:", error)
      // Mark all pending/preparing files as failed
      validFiles.forEach((file) => {
        const fileEntry = fileEntries.find(entry => entry.file === file)
        if (fileEntry && ['pending', 'hashing', 'preparing'].includes(fileEntry.status)) {
          updateFileStatus(fileEntry.id, { 
            status: 'failed', 
            error: error.message || 'Batch upload failed' 
          })
        }
      })
    }
  }

  // Legacy single file support (now uses batch internally)
  const processFile = async (file: File) => {
    await processFiles([file])
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
    processFiles,
    removeFile,
    clearCompleted,
    clearAll,
    isUploading: uploadingFiles.some(file => file.status === 'uploading'),
    hasErrors: uploadingFiles.some(file => file.status === 'failed'),
    batchId,
  }
}
