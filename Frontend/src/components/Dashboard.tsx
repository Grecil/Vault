import { useState, useCallback, useEffect, useRef } from 'react'
import { useDropzone } from 'react-dropzone'
import { formatFileSize } from '../utils/crypto'
import { useDarkMode } from '../hooks/useDarkMode'
import { useFileUpload } from '../hooks/useFileUpload'
import { useFiles } from '../hooks/useFiles'
import { useFileSearch } from '../hooks/useFileSearch'
import { useFileSort } from '../hooks/useFileSort'
import { useUser, useAuth } from '@clerk/clerk-react'
import { useToast } from '../hooks/useToast'
import { useStorageStatsRefreshTrigger } from '../hooks/useStorageStatsRefresh'
import DashboardSidebar, { type SidebarItem } from './DashboardSidebar'
import FilesView from './FilesView'
import SharedFilesView from './SharedFilesView'
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
  const [isMobileSidebarOpen, setIsMobileSidebarOpen] = useState(false)
  
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
            return isNaN(date.getTime()) ? 'Unknown' : date.toLocaleString()
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

  // Initialize fuzzy search
  const {
    searchQuery,
    setSearchQuery,
    searchResults,
    clearSearch
  } = useFileSearch(fileItems)

  // Initialize sorting
  const {
    sortField,
    sortOrder,
    sortedFiles,
    handleSort
  } = useFileSort(searchResults)

  const renderMainContent = () => {
    switch (activeTab) {
      case 'files':
        return (
          <FilesView
            files={sortedFiles}
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
            sortField={sortField}
            sortOrder={sortOrder}
            onSort={handleSort}
          />
        )
      case 'shared':
        return (
          <SharedFilesView
            files={fileItems}
            originalFiles={files}
            onCopyShareLink={async (fileId: string) => {
              try {
                await copyShareLink(fileId)
                success('Link copied!', 'Share link has been copied to your clipboard')
              } catch (err: any) {
                showError('Copy failed', err.message || 'Failed to copy share link')
              }
            }}
            onToggleFileVisibility={async (fileId: string) => {
              try {
                await toggleFileVisibility(fileId)
                success('File made private', 'Share link is no longer accessible')
              } catch (err: any) {
                showError('Update failed', err.message || 'Failed to make file private')
              }
            }}
          />
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
        {/* Mobile Sidebar Overlay */}
        {isMobileSidebarOpen && (
          <div 
            className="fixed inset-0 bg-black bg-opacity-50 z-40 lg:hidden"
            onClick={() => setIsMobileSidebarOpen(false)}
          />
        )}

        {/* Sidebar */}
        <div className={`${
          isMobileSidebarOpen ? 'translate-x-0' : '-translate-x-full'
        } fixed inset-y-0 left-0 z-50 w-64 transition-transform duration-300 ease-in-out lg:translate-x-0 lg:static lg:inset-0`}>
          <DashboardSidebar
            sidebarItems={sidebarItems}
            activeTab={activeTab}
            onTabChange={(tabId) => {
              handleTabChange(tabId)
              setIsMobileSidebarOpen(false) // Close mobile sidebar when tab changes
            }}
            fileCount={totalCount}
          />
        </div>

        {/* Main Content */}
        <div className="flex-1 flex flex-col overflow-hidden min-w-0">
          {/* Header */}
          <header className="bg-card border-b border-border px-4 sm:px-6 py-4">
            <div className="flex justify-between items-center">
              <div className="flex items-center space-x-2 sm:space-x-4 min-w-0">
                {/* Mobile menu button */}
                <button
                  onClick={() => setIsMobileSidebarOpen(true)}
                  className="p-2 rounded-lg border border-border bg-card hover:bg-muted text-foreground transition-colors lg:hidden"
                  aria-label="Open sidebar"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                  </svg>
                </button>
                
                <div className="min-w-0">
                  <h2 className="text-lg font-semibold text-foreground">Dashboard</h2>
                  {user && (
                    <span className="text-xs sm:text-sm text-muted-foreground truncate block">
                      Welcome, {user.firstName || user.emailAddresses?.[0]?.emailAddress || 'User'}
                    </span>
                  )}
                </div>
              </div>
              
              <div className="flex items-center space-x-2 sm:space-x-4">
                {/* Search - Hidden on mobile, shown on larger screens */}
                <div className="relative hidden sm:block">
                  <input
                    type="text"
                    placeholder="Search files..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="bg-muted border border-border rounded-lg px-3 py-1 text-sm text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring w-32 md:w-48"
                  />
                  {searchQuery ? (
                    <button
                      onClick={clearSearch}
                      className="absolute right-2 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                    >
                      <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  ) : (
                    <SearchIcon className="w-3 h-3 absolute right-2 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
                  )}
                </div>
                
                {/* Dark Mode Toggle */}
                <button
                  onClick={toggleDarkMode}
                  className="p-2 rounded-lg border border-border bg-card hover:bg-muted transition-colors"
                  aria-label="Toggle dark mode"
                >
                  {isDarkMode ? <SunIcon className="w-4 h-4 text-foreground" /> : <MoonIcon className="w-4 h-4 text-foreground" />}
                </button>
                
                {/* Logout Button - Text hidden on mobile */}
                <button
                  onClick={() => signOut()}
                  className="px-2 py-2 sm:px-4 rounded-lg border border-border bg-card hover:bg-muted transition-colors text-sm font-medium text-foreground"
                  aria-label="Sign out"
                >
                  <span className="hidden sm:inline">Sign Out</span>
                  <svg className="w-4 h-4 sm:hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                  </svg>
                </button>
              </div>
            </div>
          </header>

          {/* Main Content Area */}
          <main className="flex-1 overflow-auto p-4 sm:p-6">
            {renderMainContent()}
          </main>
        </div>
      </div>
    </>
  )
}

export default Dashboard
