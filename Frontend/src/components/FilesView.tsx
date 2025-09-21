import React from 'react'
import { useDropzone } from 'react-dropzone'
import { formatFileSize } from '../utils/crypto'
import { UploadIcon, GridIcon, ListIcon } from './FileTypeIcons'
import FileGrid, { FileItem } from './FileGrid'
import FileList from './FileList'
import UploadProgress from './UploadProgress'
import { UploadingFile } from '../hooks/useFileUpload'

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
  disabled = false
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
    console.log('File clicked:', file)
    // TODO: Handle file click (preview, download, etc.)
  }

  const handleMoreClick = (file: FileItem) => {
    console.log('More clicked for file:', file)
    // TODO: Show context menu or file actions
  }

  return (
    <div {...getRootProps()} className={getDropzoneClassName()}>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold text-foreground">My Files</h1>
            <p className="text-muted-foreground">
              {isDragActive 
                ? "Drop files here to upload..." 
                : "Manage and organize your files • Drag & drop to upload"
              }
            </p>
          </div>
          <div className="flex items-center space-x-3">
            {/* Upload Button */}
            <label className="bg-primary text-primary-foreground px-4 py-2 rounded-lg hover:opacity-90 transition-opacity flex items-center space-x-2 cursor-pointer">
              <input {...getInputProps()} />
              <UploadIcon />
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
              >
                <GridIcon />
              </button>
              <button
                onClick={() => onViewModeChange('list')}
                className={`p-2 rounded-md transition-colors ${
                  viewMode === 'list' 
                    ? 'bg-background text-foreground shadow-sm' 
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <ListIcon />
              </button>
            </div>
          </div>
        </div>

        {/* Upload Progress */}
        <UploadProgress 
          uploadingFiles={uploadingFiles}
          onRemoveFile={onRemoveUploadingFile}
          onClearCompleted={onClearCompleted}
          onClearAll={onClearAll}
        />

        {/* Files Display */}
        {viewMode === 'grid' ? (
          <FileGrid files={files} onFileClick={handleFileClick} />
        ) : (
          <FileList 
            files={files} 
            onFileClick={handleFileClick}
            onMoreClick={handleMoreClick}
          />
        )}

        {/* Empty State */}
        {files.length === 0 && uploadingFiles.length === 0 && (
          <div className="text-center py-12">
            <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
              <UploadIcon className="w-8 h-8 text-muted-foreground" />
            </div>
            <h3 className="text-lg font-medium text-foreground mb-2">No files yet</h3>
            <p className="text-muted-foreground mb-4">
              Get started by uploading your first file
            </p>
            <label className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:opacity-90 transition-opacity cursor-pointer inline-flex items-center space-x-2">
              <input {...getInputProps()} />
              <UploadIcon />
              <span>Upload Files</span>
            </label>
          </div>
        )}
      </div>
    </div>
  )
}

export default FilesView
