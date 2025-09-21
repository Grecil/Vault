import React from 'react'
import { getFileIcon, MoreIcon } from './FileTypeIcons'
import { type FileItem } from './FileGrid'

interface FileListProps {
  files: FileItem[]
  onFileClick?: (file: FileItem) => void
  onMoreClick?: (file: FileItem) => void
  onFileDelete?: (file: FileItem) => void
  onToggleVisibility?: (file: FileItem) => void
}

const FileList: React.FC<FileListProps> = ({ files, onFileClick, onMoreClick, onFileDelete, onToggleVisibility }) => {
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
                    <span 
                      className="text-foreground font-medium max-w-xs break-words line-clamp-2" 
                      title={file.name}
                      style={{ wordBreak: 'break-word', hyphens: 'auto' }}
                    >
                      {file.name}
                    </span>
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
                  <div className="flex items-center space-x-2">
                    {onToggleVisibility && (
                      <button 
                        className="text-foreground hover:text-foreground hover:bg-accent p-1 rounded transition-colors"
                        onClick={(e) => {
                          e.stopPropagation()
                          onToggleVisibility(file)
                        }}
                        title={file.isPublic ? "Make private" : "Make public"}
                      >
                         {file.isPublic ? (
                           <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                             <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21" />
                           </svg>
                         ) : (
                           <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                             <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                             <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                           </svg>
                         )}
                      </button>
                    )}
                    {onFileDelete && (
                      <button 
                        className="text-destructive hover:text-destructive-foreground hover:bg-destructive p-1 rounded transition-colors"
                        onClick={(e) => {
                          e.stopPropagation()
                          onFileDelete(file)
                        }}
                        title="Delete file"
                      >
                         <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                           <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                         </svg>
                      </button>
                    )}
                    <button 
                      className="text-foreground hover:text-foreground hover:bg-accent p-1 rounded transition-colors"
                      onClick={(e) => {
                        e.stopPropagation()
                        onMoreClick?.(file)
                      }}
                    >
                      <MoreIcon />
                    </button>
                  </div>
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
