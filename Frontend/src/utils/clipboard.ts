// Utility functions for clipboard operations

/**
 * Copy text to clipboard using the modern Clipboard API with fallback
 * @param text The text to copy to clipboard
 * @returns Promise that resolves when copy is successful
 */
export const copyToClipboard = async (text: string): Promise<void> => {
  // Try modern Clipboard API first
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text)
      return
    } catch (err) {
      console.warn('Modern clipboard API failed, falling back to execCommand:', err)
    }
  }

  // Fallback to legacy method
  return new Promise((resolve, reject) => {
    try {
      // Create a temporary textarea element
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      
      textArea.focus()
      textArea.select()
      
      const success = document.execCommand('copy')
      document.body.removeChild(textArea)
      
      if (success) {
        resolve()
      } else {
        reject(new Error('Failed to copy text to clipboard'))
      }
    } catch (err) {
      reject(err)
    }
  })
}

/**
 * Check if clipboard API is available
 * @returns boolean indicating if clipboard operations are supported
 */
export const isClipboardSupported = (): boolean => {
  return !!(navigator.clipboard || document.execCommand)
}

/**
 * Generate a full share URL from a relative share link
 * @param shareLink The relative share link (e.g., "/share/abc123")
 * @returns Full URL for sharing
 */
export const getFullShareUrl = (shareLink: string): string => {
  // Extract base domain from API URL (remove /api/v1 if present)
  const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
  const baseUrl = apiUrl.replace(/\/api\/v1$/, '')
  
  // If shareLink starts with /, it's a relative path from base domain
  if (shareLink.startsWith('/')) {
    return `${baseUrl}${shareLink}`
  }
  
  // Otherwise return as-is (shouldn't happen with current implementation)
  return shareLink
}
