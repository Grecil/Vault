import React from 'react'
import { useDropzone } from 'react-dropzone'
import { UploadIcon, GridIcon, ListIcon } from './FileTypeIcons'
import FileGrid, { type FileItem } from './FileGrid'
import FileList from './FileList'
import UploadProgress from './UploadProgress'
import { type UploadingFile } from '../hooks/useFileUpload'

interface FilesViewProps {
  files: FileItem[]
  viewMode: 'grid' | 'list'
  onViewModeChange: (mode: 'grid' | 'list') => void
  uploadingFiles: UploadingFile[]
  onDrop: (acceptedFiles: File[], rejectedFiles: any[]) => void
  onRemoveUploadingFile?: (fileId: string) => void
  onClearCompleted?: () => void
  onClearAll?: () => void
  isDragActive: boolean
  isDragReject: boolean
  maxFileSize: number
  maxFiles: number
  disabled?: boolean
  loading?: boolean
  error?: string | null
  onFileDelete?: (fileId: string) => Promise<void>
  onFileToggleVisibility?: (fileId: string) => Promise<void>
  onFileDownload?: (fileId: string) => Promise<void>
}

const FilesView: React.FC<FilesViewProps> = ({
  files,
  viewMode,
  onViewModeChange,
  uploadingFiles,
  onDrop,
  onRemoveUploadingFile,
  onClearCompleted,
  onClearAll,
  isDragActive,
  isDragReject,
  maxFileSize,
  maxFiles,
  disabled = false,
  loading = false,
  error = null,
  onFileDelete,
  onFileToggleVisibility,
  onFileDownload
}) => {
  const { getRootProps, getInputProps } = useDropzone({
    onDrop,
    maxFiles: maxFiles - uploadingFiles.length,
    maxSize: maxFileSize,
    multiple: true,
    disabled,
    noClick: true, // Disable click to open file picker - upload button will handle this
  })

  const getDropzoneClassName = () => {
    let baseClass = "min-h-full transition-colors duration-200"
    
    if (disabled) return baseClass
    
    if (isDragReject) {
      baseClass += " bg-destructive/5 border-destructive/20"
    } else if (isDragActive) {
      baseClass += " bg-primary/5 border-primary/20"
    }
    
    return baseClass
  }

  const handleFileClick = (file: FileItem) => {
    if (onFileDownload) {
      onFileDownload(file.id)
    }
  }

  const handleMoreClick = (file: FileItem) => {
    console.log('More clicked for file:', file)
    // TODO: Show context menu or file actions
  }

  const handleFileDelete = async (file: FileItem) => {
    if (onFileDelete && confirm(`Are you sure you want to delete "${file.name}"?`)) {
      try {
        await onFileDelete(file.id)
      } catch (error) {
        console.error('Error deleting file:', error)
      }
    }
  }

  const handleToggleVisibility = async (file: FileItem) => {
    if (onFileToggleVisibility) {
      try {
        await onFileToggleVisibility(file.id)
      } catch (error) {
        console.error('Error toggling file visibility:', error)
      }
    }
  }

  return (
    <div {...getRootProps()} className={getDropzoneClassName()}>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center space-y-4 sm:space-y-0">
          <div className="min-w-0">
            <h1 className="text-xl sm:text-2xl font-bold text-foreground">My Files</h1>
            <p className="text-sm sm:text-base text-muted-foreground">
              {isDragActive 
                ? "Drop files here to upload..." 
                : "Manage and organize your files • Drag & drop to upload"
              }
            </p>
          </div>
          <div className="flex items-center justify-between sm:justify-end space-x-3">
            {/* Upload Button */}
            <label className="bg-primary text-primary-foreground px-3 sm:px-4 py-2 rounded-lg hover:opacity-90 transition-opacity flex items-center space-x-2 cursor-pointer text-sm sm:text-base">
              <input {...getInputProps()} />
              <UploadIcon className="w-4 h-4" />
              <span>Upload</span>
            </label>
            
            {/* View Toggle */}
            <div className="flex bg-muted rounded-lg p-1">
              <button
                onClick={() => onViewModeChange('grid')}
                className={`p-2 rounded-md transition-colors ${
                  viewMode === 'grid' 
                    ? 'bg-background text-foreground shadow-sm' 
                    : 'text-muted-foreground hover:text-foreground'
                }`}
                aria-label="Grid view"
              >
                <GridIcon className="w-4 h-4" />
              </button>
              <button
                onClick={() => onViewModeChange('list')}
                className={`p-2 rounded-md transition-colors ${
                  viewMode === 'list' 
                    ? 'bg-background text-foreground shadow-sm' 
                    : 'text-muted-foreground hover:text-foreground'
                }`}
                aria-label="List view"
              >
                <ListIcon className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        {/* Error Display */}
        {error && (
          <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4">
            <p className="text-destructive font-medium">Error: {error}</p>
          </div>
        )}

        {/* Upload Progress */}
        <UploadProgress 
          uploadingFiles={uploadingFiles}
          onRemoveFile={onRemoveUploadingFile}
          onClearCompleted={onClearCompleted}
          onClearAll={onClearAll}
        />

        {/* Loading State */}
        {loading && !files.length && (
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">Loading files...</p>
          </div>
        )}

        {/* Files Display */}
        {!loading && (
          viewMode === 'grid' ? (
            <FileGrid 
              files={files} 
              onFileClick={handleFileClick}
              onFileDelete={handleFileDelete}
              onToggleVisibility={handleToggleVisibility}
            />
          ) : (
            <FileList 
              files={files} 
              onFileClick={handleFileClick}
              onMoreClick={handleMoreClick}
              onFileDelete={handleFileDelete}
              onToggleVisibility={handleToggleVisibility}
            />
          )
        )}

        {/* Empty State */}
        {files.length === 0 && uploadingFiles.length === 0 && (
          <div className="text-center py-8 sm:py-12">
            <div className="w-12 h-12 sm:w-16 sm:h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
              <UploadIcon className="w-6 h-6 sm:w-8 sm:h-8 text-muted-foreground" />
            </div>
            <h3 className="text-base sm:text-lg font-medium text-foreground mb-2">No files yet</h3>
            <p className="text-sm sm:text-base text-muted-foreground mb-4">
              Get started by uploading your first file
            </p>
            <label className="bg-primary text-primary-foreground px-4 sm:px-6 py-2 sm:py-3 rounded-lg hover:opacity-90 transition-opacity cursor-pointer inline-flex items-center space-x-2 text-sm sm:text-base">
              <input {...getInputProps()} />
              <UploadIcon className="w-4 h-4" />
              <span>Upload Files</span>
            </label>
          </div>
        )}
      </div>
    </div>
  )
}

export default FilesView
