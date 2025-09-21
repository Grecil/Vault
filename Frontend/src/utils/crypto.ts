/**
 * Calculate SHA256 hash of a file using Web Crypto API
 */
export async function calculateSHA256(file: File): Promise<string> {
  const arrayBuffer = await file.arrayBuffer()
  const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
  return hashHex
}

/**
 * Format file size in human readable format
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes'
  
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * Get file type category from MIME type
 */
export function getFileTypeCategory(mimeType: string): string {
  if (mimeType.startsWith('image/')) return 'image'
  if (mimeType.startsWith('video/')) return 'video'
  if (mimeType.startsWith('audio/')) return 'audio'
  if (mimeType === 'application/pdf') return 'pdf'
  if (mimeType.includes('spreadsheet') || mimeType.includes('excel')) return 'spreadsheet'
  if (mimeType.includes('presentation') || mimeType.includes('powerpoint')) return 'presentation'
  if (mimeType.startsWith('text/')) return 'text'
  return 'document'
}

/**
 * Validate file before upload
 */
export function validateFile(file: File, maxSizeBytes: number = 10 * 1024 * 1024) {
  const errors: string[] = []
  
  // Check file size
  if (file.size > maxSizeBytes) {
    errors.push(`File size (${formatFileSize(file.size)}) exceeds limit (${formatFileSize(maxSizeBytes)})`)
  }
  
  // Check if file is empty
  if (file.size === 0) {
    errors.push('File is empty')
  }
  
  // Basic filename validation
  if (!file.name || file.name.trim() === '') {
    errors.push('File must have a name')
  }
  
  return {
    isValid: errors.length === 0,
    errors
  }
}
