import React from 'react'
import { getFileIcon } from './FileTypeIcons'
import { type FileItem } from './FileGrid'

interface SharedFilesViewProps {
  files: FileItem[]
  originalFiles: any[] // Original file objects with download count
  onCopyShareLink: (fileId: string) => Promise<void>
  onToggleFileVisibility: (fileId: string) => Promise<void>
}

const SharedFilesView: React.FC<SharedFilesViewProps> = ({
  files,
  originalFiles,
  onCopyShareLink,
  onToggleFileVisibility
}) => {
  const sharedFiles = files?.filter(f => f?.isPublic) || []

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl sm:text-2xl font-bold text-foreground">Shared Files</h1>
        <p className="text-sm sm:text-base text-muted-foreground">Files you've shared publicly ({sharedFiles.length} files)</p>
      </div>
      
      {sharedFiles.length === 0 ? (
        <div className="bg-card border border-border rounded-lg p-6 sm:p-8 text-center">
          <svg className="w-12 h-12 sm:w-16 sm:h-16 text-muted-foreground mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
          </svg>
          <p className="text-foreground font-medium mb-2">No shared files</p>
          <p className="text-sm sm:text-base text-muted-foreground">Make files public to share them with others</p>
        </div>
      ) : (
        <div className="space-y-4">
          {sharedFiles.map((file) => (
            <div key={file.id} className="bg-card border border-border rounded-lg p-4 sm:p-6">
              <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between space-y-4 sm:space-y-0">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-3 mb-2">
                    <div className="flex-shrink-0">
                      {/* File type icon */}
                      {getFileIcon(file.type)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <h3 className="text-base sm:text-lg font-medium text-foreground truncate">{file.name}</h3>
                      <div className="flex flex-wrap items-center gap-1 sm:gap-4 text-xs sm:text-sm text-muted-foreground">
                        <span>{file.size}</span>
                        <span className="hidden sm:inline">•</span>
                        <span>Uploaded {file.uploadDate}</span>
                        <span className="hidden sm:inline">•</span>
                        <span>{originalFiles.find(f => f.id === file.id)?.downloadCount || 0} downloads</span>
                      </div>
                    </div>
                  </div>
                </div>
                
                <div className="flex flex-col sm:flex-row items-stretch sm:items-center space-y-2 sm:space-y-0 sm:space-x-2 sm:ml-4">
                  {/* Copy Link Button */}
                  <button
                    onClick={() => onCopyShareLink(file.id)}
                    className="inline-flex items-center justify-center px-3 py-2 border border-border rounded-md text-xs sm:text-sm font-medium text-foreground bg-background hover:bg-muted transition-colors"
                  >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                    Copy Link
                  </button>
                  
                  {/* Make Private Button */}
                  <button
                    onClick={() => {
                      if (confirm(`Are you sure you want to make "${file.name}" private? The share link will stop working and won't be accessible to anyone.`)) {
                        onToggleFileVisibility(file.id)
                      }
                    }}
                    className="inline-flex items-center justify-center px-3 py-2 border border-border rounded-md text-xs sm:text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                  >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21" />
                    </svg>
                    <span className="hidden sm:inline">Make Private</span>
                    <span className="sm:hidden">Private</span>
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default SharedFilesView
