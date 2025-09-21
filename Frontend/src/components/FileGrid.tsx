import React from 'react'
import { getFileIcon, PublicIcon, PrivateIcon } from './FileTypeIcons'

export interface FileItem {
  id: number
  name: string
  size: string
  type: string
  uploadDate: string
  isPublic: boolean
}

interface FileGridProps {
  files: FileItem[]
  onFileClick?: (file: FileItem) => void
}

const FileGrid: React.FC<FileGridProps> = ({ files, onFileClick }) => {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
      {files.map((file) => (
        <div 
          key={file.id} 
          className="bg-card border border-border rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer group"
          onClick={() => onFileClick?.(file)}
        >
          <div className="flex flex-col items-center space-y-2">
            {getFileIcon(file.type)}
            <div className="text-center">
              <p 
                className="text-sm font-medium text-foreground truncate w-full" 
                title={file.name}
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
