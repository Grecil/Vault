import React from 'react'
import { getFileIcon, PrivateIcon, PublicIcon } from './FileTypeIcons'
import { type FileItem } from './FileGrid'

interface FileListProps {
  files: FileItem[]
  onFileClick?: (file: FileItem) => void
  onMoreClick?: (file: FileItem) => void
  onFileDelete?: (file: FileItem) => void
  onToggleVisibility?: (file: FileItem) => void
}

const FileList: React.FC<FileListProps> = ({ files, onFileClick, onFileDelete, onToggleVisibility }) => {
  return (
    <div className="bg-card border border-border rounded-lg overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[600px]">
          <thead className="bg-muted border-b border-border">
            <tr>
              <th className="text-left py-2 sm:py-3 px-3 sm:px-4 font-medium text-foreground text-xs sm:text-sm">Name</th>
              <th className="text-left py-2 sm:py-3 px-3 sm:px-4 font-medium text-foreground text-xs sm:text-sm hidden sm:table-cell">Size</th>
              <th className="text-left py-2 sm:py-3 px-3 sm:px-4 font-medium text-foreground text-xs sm:text-sm hidden md:table-cell">Type</th>
              <th className="text-left py-2 sm:py-3 px-3 sm:px-4 font-medium text-foreground text-xs sm:text-sm">Visibility</th>
              <th className="text-left py-2 sm:py-3 px-3 sm:px-4 font-medium text-foreground text-xs sm:text-sm hidden lg:table-cell">Uploaded at</th>
              <th className="text-left py-2 sm:py-3 px-3 sm:px-4 font-medium text-foreground text-xs sm:text-sm">Actions</th>
            </tr>
          </thead>
          <tbody>
            {files.map((file) => (
              <tr 
                key={file.id} 
                className="border-b border-border hover:bg-muted/50 transition-colors cursor-pointer"
                onClick={() => onFileClick?.(file)}
              >
                <td className="py-2 sm:py-3 px-3 sm:px-4">
                  <div className="flex items-center space-x-2 sm:space-x-3">
                    <div className="flex-shrink-0">
                      {getFileIcon(file.type)}
                    </div>
                    <div className="min-w-0 flex-1">
                      <span 
                        className="text-foreground font-medium text-xs sm:text-sm break-words line-clamp-2 block" 
                        title={file.name}
                        style={{ wordBreak: 'break-word', hyphens: 'auto' }}
                      >
                        {file.name}
                      </span>
                      {/* Show size and type on mobile in subtitle */}
                      <div className="sm:hidden text-xs text-muted-foreground mt-1">
                        {file.size} • {file.type}
                      </div>
                    </div>
                  </div>
                </td>
                <td className="py-2 sm:py-3 px-3 sm:px-4 text-muted-foreground text-xs sm:text-sm hidden sm:table-cell">{file.size}</td>
                <td className="py-2 sm:py-3 px-3 sm:px-4 text-muted-foreground text-xs sm:text-sm capitalize hidden md:table-cell">{file.type}</td>
                <td className="py-2 sm:py-3 px-3 sm:px-4">
                  {file.isPublic ? (
                    <span className="inline-flex items-center px-2 py-1 rounded text-xs bg-chart-2/20 text-chart-2 border border-chart-2/30">
                      <span className="hidden sm:inline">Public</span>
                      <span className="sm:hidden">Pub</span>
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-2 py-1 rounded text-xs bg-muted text-muted-foreground border border-border">
                      <span className="hidden sm:inline">Private</span>
                      <span className="sm:hidden">Prv</span>
                    </span>
                  )}
                </td>
                <td className="py-2 sm:py-3 px-3 sm:px-4 text-muted-foreground text-xs sm:text-sm hidden lg:table-cell">{file.uploadDate}</td>
                <td className="py-2 sm:py-3 px-3 sm:px-4">
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
                         {file.isPublic ? <PrivateIcon className="w-4 h-4" /> : <PublicIcon className="w-4 h-4" />}
                      </button>
                    )}
                    {onFileDelete && (
                      <button 
                        className="text-primary hover:text-primary-foreground hover:bg-primary p-1 rounded transition-colors"
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
