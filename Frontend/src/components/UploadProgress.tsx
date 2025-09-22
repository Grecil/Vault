import React from 'react'
import type { UploadingFile } from '../hooks/useFileUpload'

interface UploadProgressProps {
  uploadingFiles: UploadingFile[]
  onRemoveFile?: (fileId: string) => void
  onClearCompleted?: () => void
  onClearAll?: () => void
}

const UploadProgress: React.FC<UploadProgressProps> = ({ 
  uploadingFiles, 
  onRemoveFile,
  onClearCompleted,
  onClearAll 
}) => {
  if (uploadingFiles.length === 0) return null

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'text-chart-2'
      case 'failed': return 'text-destructive'
      case 'uploading': return 'text-primary'
      case 'hashing': return 'text-muted-foreground'
      case 'preparing': return 'text-secondary'
      case 'duplicate': return 'text-chart-4'
      default: return 'text-muted-foreground'
    }
  }

  const getStatusText = (fileEntry: UploadingFile) => {
    if (fileEntry.status === 'uploading') {
      return `${Math.round(fileEntry.progress)}%`
    }
    return fileEntry.status.charAt(0).toUpperCase() + fileEntry.status.slice(1)
  }

  const completedCount = uploadingFiles.filter(f => f.status === 'completed').length
  const failedCount = uploadingFiles.filter(f => f.status === 'failed').length

  return (
    <div className="bg-card border border-border rounded-lg p-3 sm:p-4 space-y-3">
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center space-y-2 sm:space-y-0">
        <h3 className="text-base sm:text-lg font-semibold text-foreground">Upload Progress</h3>
        <div className="flex flex-wrap gap-2">
          {completedCount > 0 && onClearCompleted && (
            <button
              onClick={onClearCompleted}
              className="text-xs text-muted-foreground hover:text-foreground px-2 py-1 rounded border border-border hover:bg-muted transition-colors whitespace-nowrap"
            >
              Clear Completed ({completedCount})
            </button>
          )}
          {uploadingFiles.length > 0 && onClearAll && (
            <button
              onClick={onClearAll}
              className="text-xs text-muted-foreground hover:text-foreground px-2 py-1 rounded border border-border hover:bg-muted transition-colors whitespace-nowrap"
            >
              Clear All
            </button>
          )}
        </div>
      </div>

      <div className="space-y-3 max-h-48 overflow-y-auto">
        {uploadingFiles.map((fileEntry) => (
          <div key={fileEntry.id} className="flex items-start space-x-2 sm:space-x-3">
            <div className="flex-1 min-w-0">
              <div className="flex justify-between items-start">
                <p className="text-xs sm:text-sm font-medium text-foreground break-words line-clamp-2 flex-1 pr-2" title={fileEntry.file.name}>
                  {fileEntry.file.name}
                </p>
                <div className="flex items-center space-x-1 sm:space-x-2 flex-shrink-0">
                  <span className={`text-xs font-semibold ${getStatusColor(fileEntry.status)} whitespace-nowrap`}>
                    {getStatusText(fileEntry)}
                  </span>
                  {(fileEntry.status === 'completed' || fileEntry.status === 'failed') && onRemoveFile && (
                    <button
                      onClick={() => onRemoveFile(fileEntry.id)}
                      className="text-xs text-muted-foreground hover:text-foreground p-1 rounded hover:bg-muted transition-colors"
                      title="Remove"
                    >
                      ✕
                    </button>
                  )}
                </div>
              </div>
              
              {fileEntry.status === 'uploading' && (
                <div className="w-full bg-muted rounded-full h-1 mt-1">
                  <div 
                    className="bg-primary h-1 rounded-full transition-all duration-300" 
                    style={{ width: `${fileEntry.progress}%` }}
                  />
                </div>
              )}
              
              {fileEntry.error && (
                <p className="text-xs text-destructive mt-1">{fileEntry.error}</p>
              )}
              
              {fileEntry.status === 'duplicate' && fileEntry.existingFile && (
                <p className="text-xs text-chart-4 mt-1">
                  Duplicate of existing file: {fileEntry.existingFile.name}
                </p>
              )}
              
              {fileEntry.status === 'completed' && (
                <p className="text-xs text-chart-2 mt-1">
                  Successfully uploaded • {(fileEntry.file.size / 1024 / 1024).toFixed(2)} MB
                </p>
              )}
            </div>
          </div>
        ))}
      </div>
      
      {/* Summary */}
      {(completedCount > 0 || failedCount > 0) && (
        <div className="text-xs text-muted-foreground pt-2 border-t border-border">
          {completedCount > 0 && <span className="text-chart-2">{completedCount} completed</span>}
          {completedCount > 0 && failedCount > 0 && <span className="mx-2">•</span>}
          {failedCount > 0 && <span className="text-destructive">{failedCount} failed</span>}
        </div>
      )}
    </div>
  )
}

export default UploadProgress
