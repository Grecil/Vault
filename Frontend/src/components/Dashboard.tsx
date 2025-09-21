import { useState, useCallback } from 'react'
import { useDropzone } from 'react-dropzone'
import { formatFileSize } from '../utils/crypto'
import { useDarkMode } from '../hooks/useDarkMode'
import { useFileUpload } from '../hooks/useFileUpload'
import DashboardSidebar, { type SidebarItem } from './DashboardSidebar'
import FilesView from './FilesView'
import { SearchIcon, SunIcon, MoonIcon } from './FileTypeIcons'

const Dashboard = () => {
  const [activeTab, setActiveTab] = useState('files')
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  
  const { isDarkMode, toggleDarkMode } = useDarkMode()
  const { 
    uploadingFiles, 
    processFile, 
    removeFile, 
    clearCompleted, 
    clearAll 
  } = useFileUpload()

  const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB

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

    // Process accepted files
    if (acceptedFiles.length > 0) {
      acceptedFiles.forEach(processFile)
    }
  }, [processFile])

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
      id: 'settings',
      name: 'Settings',
      icon: (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
      )
    }
  ]

  // Mock file data for now
  const mockFiles = [
    { id: 1, name: 'Project Proposal.pdf', size: '2.5 MB', type: 'pdf', uploadDate: '2024-01-15', isPublic: false },
    { id: 2, name: 'vacation-photo.jpg', size: '5.2 MB', type: 'image', uploadDate: '2024-01-14', isPublic: true },
    { id: 3, name: 'budget-2024.xlsx', size: '890 KB', type: 'spreadsheet', uploadDate: '2024-01-13', isPublic: false },
    { id: 4, name: 'meeting-notes.txt', size: '45 KB', type: 'text', uploadDate: '2024-01-12', isPublic: false },
    { id: 5, name: 'presentation.pptx', size: '12.3 MB', type: 'presentation', uploadDate: '2024-01-11', isPublic: true },
    { id: 6, name: 'invoice-jan-2024.pdf', size: '234 KB', type: 'pdf', uploadDate: '2024-01-10', isPublic: false }
  ]

  const renderMainContent = () => {
    switch (activeTab) {
      case 'files':
        return (
          <FilesView
            files={mockFiles}
            viewMode={viewMode}
            onViewModeChange={setViewMode}
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
          />
        )
      case 'shared':
        return (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-foreground">Shared Files</h1>
              <p className="text-muted-foreground">Files you've shared publicly</p>
            </div>
            <div className="bg-card border border-border rounded-lg p-8 text-center">
              <svg className="w-16 h-16 text-muted-foreground mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
              </svg>
              <p className="text-foreground font-medium mb-2">Shared files view</p>
              <p className="text-muted-foreground">Shows {mockFiles.filter(f => f.isPublic).length} public files</p>
            </div>
          </div>
        )
      case 'settings':
        return (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-foreground">Settings</h1>
              <p className="text-muted-foreground">Manage your account and preferences</p>
            </div>
            <div className="bg-card border border-border rounded-lg p-8 text-center">
              <svg className="w-16 h-16 text-muted-foreground mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              <p className="text-foreground font-medium mb-2">Settings panel</p>
              <p className="text-muted-foreground">Configuration options coming soon</p>
            </div>
          </div>
        )
      default:
        return null
    }
  }

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <DashboardSidebar
        sidebarItems={sidebarItems}
        activeTab={activeTab}
        onTabChange={setActiveTab}
        fileCount={mockFiles.length}
      />

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="bg-card border-b border-border px-6 py-4">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-4">
              <h2 className="text-lg font-semibold text-foreground">Dashboard</h2>
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
            </div>
          </div>
        </header>

        {/* Main Content Area */}
        <main className="flex-1 overflow-auto p-6">
          {renderMainContent()}
        </main>
      </div>
    </div>
  )
}

export default Dashboard
