import { useState, useCallback, useEffect, useRef } from 'react'
import { useDropzone } from 'react-dropzone'
import { formatFileSize } from '../utils/crypto'
import { useDarkMode } from '../hooks/useDarkMode'
import { useFileUpload } from '../hooks/useFileUpload'
import { useFiles } from '../hooks/useFiles'
import { useUser, useAuth } from '@clerk/clerk-react'
import { useToast } from '../hooks/useToast'
import { useStorageStatsRefreshTrigger } from '../hooks/useStorageStatsRefresh'
import DashboardSidebar, { type SidebarItem } from './DashboardSidebar'
import FilesView from './FilesView'
import StorageStatistics from './StorageStatistics'
import { ToastContainer } from './Toast'
import { SearchIcon, SunIcon, MoonIcon } from './FileTypeIcons'

const Dashboard = () => {
  const [activeTab, setActiveTab] = useState(() => {
    const savedTab = localStorage.getItem('activeTab')
    return (savedTab && ['files', 'shared', 'statistics'].includes(savedTab)) ? savedTab : 'files'
  })
  const [viewMode, setViewMode] = useState<'grid' | 'list'>(() => {
    const savedViewMode = localStorage.getItem('fileViewMode')
    return (savedViewMode as 'grid' | 'list') || 'grid'
  })
  
  const { isDarkMode, toggleDarkMode } = useDarkMode()
  const { 
    uploadingFiles, 
    processFiles,
    removeFile, 
    clearCompleted, 
    clearAll 
  } = useFileUpload()
  
  const {
    files,
    loading,
    error: filesError,
    refreshFiles,
    deleteFile,
    toggleFileVisibility,
    copyShareLink,
    downloadFile,
    totalCount
  } = useFiles()
  
  const { user } = useUser()
  const { signOut } = useAuth()
  const { toasts, removeToast, success, error: showError } = useToast()
  const { triggerRefresh } = useStorageStatsRefreshTrigger()

  // Track which files have already shown toast notifications
  const processedToastsRef = useRef<Set<string>>(new Set())

  const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB

  const handleViewModeChange = (mode: 'grid' | 'list') => {
    setViewMode(mode)
    localStorage.setItem('fileViewMode', mode)
  }

  const handleTabChange = (tabId: string) => {
    setActiveTab(tabId)
    localStorage.setItem('activeTab', tabId)
  }

  const onDrop = useCallback((acceptedFiles: File[], rejectedFiles: any[]) => {
    // Handle rejected files
    if (rejectedFiles.length > 0) {
      rejectedFiles.forEach((rejection) => {
        const { file, errors } = rejection
        const errorMessages = errors.map((e: any) => {
          switch (e.code) {
            case 'file-too-large':
              return `File "${file.name}" is too large (max ${formatFileSize(MAX_FILE_SIZE)})`
            case 'too-many-files':
              return `Too many files selected (max 10)`
            default:
              return `File "${file.name}" was rejected: ${e.message}`
          }
        }).join(', ')
        
        console.error('File rejected:', errorMessages)
      })
    }

    // Process accepted files as a batch
    if (acceptedFiles.length > 0) {
      processFiles(acceptedFiles)
    }
  }, [processFiles])

  // Only enable dropzone for files tab
  const { isDragActive, isDragReject } = useDropzone({
    onDrop,
    maxFiles: 10 - uploadingFiles.length,
    maxSize: MAX_FILE_SIZE,
    multiple: true,
    disabled: activeTab !== 'files' || uploadingFiles.length >= 10,
    noClick: true, // Disable click to open file picker - files section will handle this
  })

  const sidebarItems: SidebarItem[] = [
    {
      id: 'files',
      name: 'My Files',
      icon: (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
      )
    },
    {
      id: 'shared',
      name: 'Shared',
      icon: (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
        </svg>
      )
    },
    {
      id: 'statistics',
      name: 'Statistics',
      icon: (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      )
    }
  ]

  // Refresh files when upload completes and show notifications
  useEffect(() => {
    const processedToasts = processedToastsRef.current
    
    // Filter for files that haven't shown toasts yet
    const newCompletedUploads = uploadingFiles.filter(f => 
      f.status === 'completed' && !processedToasts.has(f.id)
    )
    const newFailedUploads = uploadingFiles.filter(f => 
      f.status === 'failed' && !processedToasts.has(f.id)
    )
    const newDuplicateUploads = uploadingFiles.filter(f => 
      f.status === 'duplicate' && !processedToasts.has(f.id)
    )
    
    // Handle completed uploads
    if (newCompletedUploads.length > 0) {
      refreshFiles()
      triggerRefresh()
      newCompletedUploads.forEach(upload => {
        success('File uploaded successfully', `${upload.file.name} has been uploaded`)
        processedToasts.add(upload.id)
      })
    }
    
    // Handle failed uploads
    if (newFailedUploads.length > 0) {
      newFailedUploads.forEach(upload => {
        showError('Upload failed', `Failed to upload ${upload.file.name}: ${upload.error}`)
        processedToasts.add(upload.id)
      })
    }
    
    // Handle duplicate uploads
    if (newDuplicateUploads.length > 0) {
      refreshFiles()
      triggerRefresh()
      newDuplicateUploads.forEach(upload => {
        success('File already exists', `${upload.file.name} was already in your vault`)
        processedToasts.add(upload.id)
      })
    }

    // Clean up processed toasts for files that are no longer in the uploadingFiles array
    // This prevents memory leaks and allows re-showing toasts for re-uploaded files
    const currentFileIds = new Set(uploadingFiles.map(f => f.id))
    const toastsToRemove = Array.from(processedToasts).filter(id => !currentFileIds.has(id))
    toastsToRemove.forEach(id => processedToasts.delete(id))
    
  }, [uploadingFiles, refreshFiles, success, showError, triggerRefresh])

  const getFileTypeFromMime = (mimeType: string): string => {
    if (mimeType.startsWith('image/')) return 'image'
    if (mimeType.startsWith('video/')) return 'video'
    if (mimeType.startsWith('audio/')) return 'audio'
    if (mimeType === 'application/pdf') return 'pdf'
    if (mimeType.includes('spreadsheet') || mimeType.includes('excel')) return 'spreadsheet'
    if (mimeType.includes('presentation') || mimeType.includes('powerpoint')) return 'presentation'
    if (mimeType.startsWith('text/')) return 'text'
    return 'document'
  }

  // Convert FileInfo to FileItem format expected by components
  const fileItems = (() => {
    try {
      return (files || []).map(file => ({
        id: file?.id || '',
        name: file?.name || 'Unknown File',
        size: (() => {
          try {
            return formatFileSize(file?.size || 0)
          } catch (e) {
            return '0 Bytes'
          }
        })(),
        type: getFileTypeFromMime(file?.contentType || 'application/octet-stream'),
        uploadDate: (() => {
          try {
            if (!file?.uploadDate) return 'Unknown'
            const date = new Date(file.uploadDate)
            return isNaN(date.getTime()) ? 'Unknown' : date.toLocaleDateString()
          } catch (e) {
            return 'Unknown'
          }
        })(),
        isPublic: file?.isPublic || false
      }))
    } catch (error) {
      console.error('Error processing files:', error)
      return []
    }
  })()

  const renderMainContent = () => {
    switch (activeTab) {
      case 'files':
        return (
          <FilesView
            files={fileItems}
            viewMode={viewMode}
            onViewModeChange={handleViewModeChange}
            uploadingFiles={uploadingFiles}
            onDrop={onDrop}
            onRemoveUploadingFile={removeFile}
            onClearCompleted={clearCompleted}
            onClearAll={clearAll}
            isDragActive={isDragActive}
            isDragReject={isDragReject}
            maxFileSize={MAX_FILE_SIZE}
            maxFiles={10}
            disabled={activeTab !== 'files' || uploadingFiles.length >= 10}
            loading={loading}
            error={filesError}
            onFileDelete={async (fileId: string) => {
              try {
                await deleteFile(fileId)
                success('File deleted', 'File has been permanently deleted')
              } catch (err: any) {
                showError('Delete failed', err.message || 'Failed to delete file')
              }
            }}
            onFileToggleVisibility={async (fileId: string) => {
              try {
                const result = await toggleFileVisibility(fileId)
                if (result.shareLink) {
                  success('Link copied!', 'File is now public and the share link has been copied to your clipboard')
                } else {
                success('Visibility updated', 'File visibility has been changed')
                }
              } catch (err: any) {
                showError('Update failed', err.message || 'Failed to update file visibility')
              }
            }}
            onFileDownload={async (fileId: string) => {
              try {
                await downloadFile(fileId)
              } catch (err: any) {
                showError('Download failed', err.message || 'Failed to download file')
              }
            }}
          />
        )
      case 'shared':
        const sharedFiles = fileItems?.filter(f => f?.isPublic) || []
        return (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-foreground">Shared Files</h1>
              <p className="text-muted-foreground">Files you've shared publicly ({sharedFiles.length} files)</p>
            </div>
            
            {sharedFiles.length === 0 ? (
            <div className="bg-card border border-border rounded-lg p-8 text-center">
              <svg className="w-16 h-16 text-muted-foreground mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
              </svg>
                <p className="text-foreground font-medium mb-2">No shared files</p>
                <p className="text-muted-foreground">Make files public to share them with others</p>
              </div>
            ) : (
              <div className="space-y-4">
                {sharedFiles.map((file) => (
                  <div key={file.id} className="bg-card border border-border rounded-lg p-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center space-x-3 mb-2">
                          <div className="flex-shrink-0">
                            {/* File type icon */}
                            <div className="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
                              <svg className="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                              </svg>
                            </div>
                          </div>
                          <div className="flex-1 min-w-0">
                            <h3 className="text-lg font-medium text-foreground truncate">{file.name}</h3>
                            <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                              <span>{file.size}</span>
                              <span>•</span>
                              <span>Uploaded {file.uploadDate}</span>
                              <span>•</span>
                              <span>{files.find(f => f.id === file.id)?.downloadCount || 0} downloads</span>
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      <div className="flex items-center space-x-2 ml-4">
                        {/* Copy Link Button */}
                        <button
                          onClick={async () => {
                            try {
                              // Copy share link without toggling visibility
                              await copyShareLink(file.id)
                              success('Link copied!', 'Share link has been copied to your clipboard')
                            } catch (err: any) {
                              showError('Copy failed', err.message || 'Failed to copy share link')
                            }
                          }}
                          className="inline-flex items-center px-3 py-2 border border-border rounded-md text-sm font-medium text-foreground bg-background hover:bg-muted transition-colors"
                        >
                          <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                          </svg>
                          Copy Link
                        </button>
                        
                        {/* Make Private Button */}
                        <button
                          onClick={async () => {
                            if (confirm(`Are you sure you want to make "${file.name}" private? The share link will stop working and won't be accessible to anyone.`)) {
                              try {
                                await toggleFileVisibility(file.id)
                                success('File made private', 'Share link is no longer accessible')
                              } catch (err: any) {
                                showError('Update failed', err.message || 'Failed to make file private')
                              }
                            }
                          }}
                          className="inline-flex items-center px-3 py-2 border border-border rounded-md text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                        >
                          <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21" />
                          </svg>
                          Make Private
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
            </div>
            )}
          </div>
        )
      case 'statistics':
        return <StorageStatistics />
      default:
        return null
    }
  }

  return (
    <>
      <ToastContainer toasts={toasts} onRemove={removeToast} />
      <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <DashboardSidebar
        sidebarItems={sidebarItems}
        activeTab={activeTab}
        onTabChange={handleTabChange}
        fileCount={totalCount}
      />

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="bg-card border-b border-border px-6 py-4">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-4">
              <h2 className="text-lg font-semibold text-foreground">Dashboard</h2>
              {user && (
                <span className="text-sm text-muted-foreground">
                  Welcome, {user.firstName || user.emailAddresses?.[0]?.emailAddress || 'User'}
                </span>
              )}
            </div>
            <div className="flex items-center space-x-4">
              {/* Search */}
              <div className="relative">
                <input
                  type="text"
                  placeholder="Search files..."
                  className="bg-muted border border-border rounded-lg px-3 py-1 text-sm text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                />
                <SearchIcon className="w-3 h-3 absolute right-2 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
              </div>
              
              {/* Dark Mode Toggle */}
              <button
                onClick={toggleDarkMode}
                className="p-2 rounded-lg border border-border bg-card hover:bg-muted transition-colors"
                aria-label="Toggle dark mode"
              >
                {isDarkMode ? <SunIcon className="w-4 h-4 text-foreground" /> : <MoonIcon className="w-4 h-4 text-foreground" />}
              </button>
              
              {/* Logout Button */}
              <button
                onClick={() => signOut()}
                className="px-4 py-2 rounded-lg border border-border bg-card hover:bg-muted transition-colors text-sm font-medium text-foreground"
                aria-label="Sign out"
              >
                Sign Out
              </button>
            </div>
          </div>
        </header>

        {/* Main Content Area */}
        <main className="flex-1 overflow-auto p-6">
          {renderMainContent()}
        </main>
      </div>
    </div>
    </>
  )
}

export default Dashboard
