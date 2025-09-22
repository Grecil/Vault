import { useState, useEffect, useCallback } from 'react'
import { useAuth } from '@clerk/clerk-react'
import { apiClient, type FileInfo } from '../lib/api'
import { useStorageStatsRefreshTrigger } from './useStorageStatsRefresh'
import { copyToClipboard, getFullShareUrl } from '../utils/clipboard'

export interface UseFilesReturn {
  files: FileInfo[]
  loading: boolean
  error: string | null
  refreshFiles: () => Promise<void>
  deleteFile: (fileId: string) => Promise<void>
  toggleFileVisibility: (fileId: string) => Promise<{ shareLink?: string }>
  copyShareLink: (fileId: string) => Promise<string>
  downloadFile: (fileId: string) => Promise<void>
  totalCount: number
  hasMore: boolean
}

export const useFiles = (): UseFilesReturn => {
  const [files, setFiles] = useState<FileInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [totalCount, setTotalCount] = useState(0)
  const [hasMore, setHasMore] = useState(false)
  const { getToken, isSignedIn } = useAuth()
  const { triggerRefresh } = useStorageStatsRefreshTrigger()

  const refreshFiles = useCallback(async () => {
    if (!isSignedIn) return
    
    setLoading(true)
    setError(null)
    
    try {
      const response = await apiClient.getUserFiles(getToken, 50, 1)
      setFiles(response.files)
      setTotalCount(response.totalCount)
      setHasMore(response.hasMore)
    } catch (err: any) {
      setError(err.message || 'Failed to load files')
      console.error('Error loading files:', err)
    } finally {
      setLoading(false)
    }
  }, [isSignedIn])

  const deleteFile = useCallback(async (fileId: string) => {
    if (!isSignedIn) return
    
    try {
      await apiClient.deleteFile(getToken, fileId)
      // Remove file from local state
      setFiles(prev => prev.filter(file => file.id !== fileId))
      setTotalCount(prev => Math.max(0, prev - 1))
      // Trigger storage stats refresh
      triggerRefresh()
    } catch (err: any) {
      setError(err.message || 'Failed to delete file')
      console.error('Error deleting file:', err)
      throw err
    }
  }, [isSignedIn, triggerRefresh])

  const toggleFileVisibility = useCallback(async (fileId: string) => {
    if (!isSignedIn) return { shareLink: undefined }
    
    // Get the current file to know what to toggle to
    const currentFile = files.find(file => file.id === fileId)
    if (!currentFile) return { shareLink: undefined }
    
    const newVisibility = !currentFile.isPublic
    
    // Optimistic update - update UI immediately
    setFiles(prev => prev.map(file => 
      file.id === fileId 
        ? { ...file, isPublic: newVisibility }
        : file
    ))
    
    try {
      const response = await apiClient.updateFileVisibility(getToken, fileId)
      
      // Handle share link
      let shareLink: string | undefined
      if (response.shareLink && newVisibility) {
        shareLink = getFullShareUrl(response.shareLink)
        // Automatically copy to clipboard
        try {
          await copyToClipboard(shareLink)
        } catch (clipboardErr) {
          console.warn('Failed to copy share link to clipboard:', clipboardErr)
          // Don't throw error for clipboard failure
        }
      }
      
      return { shareLink }
    } catch (err: any) {
      // Revert optimistic update on error
      setFiles(prev => prev.map(file => 
        file.id === fileId 
          ? { ...file, isPublic: currentFile.isPublic }
          : file
      ))
      setError(err.message || 'Failed to update file visibility')
      console.error('Error updating file visibility:', err)
      throw err
    }
  }, [isSignedIn, files, getToken])

  const copyShareLink = useCallback(async (fileId: string): Promise<string> => {
    if (!isSignedIn) throw new Error('Not signed in')
    
    // Get the current file to ensure it's public
    const currentFile = files.find(file => file.id === fileId)
    if (!currentFile) throw new Error('File not found')
    if (!currentFile.isPublic) throw new Error('File is not public')
    
    try {
      // Get share link without toggling visibility
      const response = await apiClient.getShareLink(getToken, fileId)
      
      if (response.share_link) {
        const shareLink = getFullShareUrl(response.share_link)
        await copyToClipboard(shareLink)
        return shareLink
      }
      
      throw new Error('No share link available')
    } catch (err: any) {
      setError(err.message || 'Failed to copy share link')
      console.error('Error copying share link:', err)
      throw err
    }
  }, [isSignedIn, files, getToken])

  const downloadFile = useCallback(async (fileId: string) => {
    if (!isSignedIn) return
    
    try {
      const response = await apiClient.getDownloadUrl(getToken, fileId)
      // Open download URL in new tab
      window.open(response.url, '_blank')
    } catch (err: any) {
      setError(err.message || 'Failed to download file')
      console.error('Error downloading file:', err)
      throw err
    }
  }, [isSignedIn])

  // Load files on mount and when auth becomes available
  useEffect(() => {
    if (isSignedIn) {
      refreshFiles()
    }
  }, [isSignedIn, refreshFiles])

  return {
    files,
    loading,
    error,
    refreshFiles,
    deleteFile,
    toggleFileVisibility,
    copyShareLink,
    downloadFile,
    totalCount,
    hasMore
  }
}
