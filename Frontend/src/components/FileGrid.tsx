import React from 'react'
import { getFileIcon, PublicIcon, PrivateIcon } from './FileTypeIcons'

export interface FileItem {
  id: string
  name: string
  size: string
  type: string
  uploadDate: string
  isPublic: boolean
}

interface FileGridProps {
  files: FileItem[]
  onFileClick?: (file: FileItem) => void
  onFileDelete?: (file: FileItem) => void
  onToggleVisibility?: (file: FileItem) => void
}

const FileGrid: React.FC<FileGridProps> = ({ files, onFileClick, onFileDelete, onToggleVisibility }) => {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
      {files.map((file) => (
        <div 
          key={file.id} 
          className="bg-card border border-border rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer group relative"
          onClick={() => onFileClick?.(file)}
        >
          {/* Action buttons - always show but with better contrast */}
          <div className="absolute top-2 right-2 flex space-x-1">
            {onToggleVisibility && (
              <button 
                onClick={(e) => {
                  e.stopPropagation()
                  onToggleVisibility(file)
                }}
                className="p-1 bg-background/90 backdrop-blur-sm rounded text-xs text-foreground hover:bg-accent shadow-sm border border-border/50"
                title={file.isPublic ? "Make private" : "Make public"}
              >
                {file.isPublic ? <PrivateIcon className="w-3 h-3" /> : <PublicIcon className="w-3 h-3" />}
              </button>
            )}
            {onFileDelete && (
              <button 
                onClick={(e) => {
                  e.stopPropagation()
                  onFileDelete(file)
                }}
                className="p-1 bg-background/90 backdrop-blur-sm rounded text-xs text-destructive hover:bg-destructive hover:text-destructive-foreground shadow-sm border border-border/50"
                title="Delete file"
              >
                 <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                   <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                 </svg>
              </button>
            )}
          </div>

          <div className="flex flex-col items-center space-y-2">
            {getFileIcon(file.type)}
            <div className="text-center">
              <p 
                className="text-sm font-medium text-foreground break-words line-clamp-2 w-full max-w-full px-1" 
                title={file.name}
                style={{ wordBreak: 'break-word', hyphens: 'auto' }}
              >
                {file.name}
              </p>
              <p className="text-xs text-muted-foreground">{file.size}</p>
              <div className="flex items-center justify-center mt-1">
                {file.isPublic ? (
                  <span className="inline-flex items-center px-2 py-1 rounded text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                    <PublicIcon className="mr-1" />
                    Public
                  </span>
                ) : (
                  <span className="inline-flex items-center px-2 py-1 rounded text-xs bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200">
                    <PrivateIcon className="mr-1" />
                    Private
                  </span>
                )}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}

export default FileGrid
