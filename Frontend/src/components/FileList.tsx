import React from 'react'
import { getFileIcon, MoreIcon } from './FileTypeIcons'
import { FileItem } from './FileGrid'

interface FileListProps {
  files: FileItem[]
  onFileClick?: (file: FileItem) => void
  onMoreClick?: (file: FileItem) => void
}

const FileList: React.FC<FileListProps> = ({ files, onFileClick, onMoreClick }) => {
  return (
    <div className="bg-card border border-border rounded-lg overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-muted border-b border-border">
            <tr>
              <th className="text-left py-3 px-4 font-medium text-foreground">Name</th>
              <th className="text-left py-3 px-4 font-medium text-foreground">Size</th>
              <th className="text-left py-3 px-4 font-medium text-foreground">Type</th>
              <th className="text-left py-3 px-4 font-medium text-foreground">Visibility</th>
              <th className="text-left py-3 px-4 font-medium text-foreground">Date</th>
              <th className="text-left py-3 px-4 font-medium text-foreground">Actions</th>
            </tr>
          </thead>
          <tbody>
            {files.map((file) => (
              <tr 
                key={file.id} 
                className="border-b border-border hover:bg-muted/50 transition-colors cursor-pointer"
                onClick={() => onFileClick?.(file)}
              >
                <td className="py-3 px-4">
                  <div className="flex items-center space-x-3">
                    {getFileIcon(file.type)}
                    <span className="text-foreground font-medium">{file.name}</span>
                  </div>
                </td>
                <td className="py-3 px-4 text-muted-foreground">{file.size}</td>
                <td className="py-3 px-4 text-muted-foreground capitalize">{file.type}</td>
                <td className="py-3 px-4">
                  {file.isPublic ? (
                    <span className="inline-flex items-center px-2 py-1 rounded text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                      Public
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-2 py-1 rounded text-xs bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200">
                      Private
                    </span>
                  )}
                </td>
                <td className="py-3 px-4 text-muted-foreground">{file.uploadDate}</td>
                <td className="py-3 px-4">
                  <button 
                    className="text-muted-foreground hover:text-foreground p-1 rounded"
                    onClick={(e) => {
                      e.stopPropagation()
                      onMoreClick?.(file)
                    }}
                  >
                    <MoreIcon />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default FileList
