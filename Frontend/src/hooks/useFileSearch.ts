import { useState, useMemo } from 'react'
import Fuse from 'fuse.js'
import { type FileItem } from '../components/FileGrid'

interface UseFileSearchOptions {
  threshold?: number // 0.0 = exact match, 1.0 = match anything
  minMatchCharLength?: number
  includeScore?: boolean
  includeMatches?: boolean
}

interface UseFileSearchReturn {
  searchQuery: string
  setSearchQuery: (query: string) => void
  searchResults: FileItem[]
  clearSearch: () => void
  isSearching: boolean
}

const defaultOptions: UseFileSearchOptions = {
  threshold: 0.6, // More lenient threshold for better matches
  minMatchCharLength: 1,
  includeScore: true,
  includeMatches: true
}

export const useFileSearch = (
  files: FileItem[], 
  options: UseFileSearchOptions = {}
): UseFileSearchReturn => {
  const [searchQuery, setSearchQuery] = useState('')
  
  // Merge options with defaults
  const searchOptions = { ...defaultOptions, ...options }
  
  // Fuse configuration
  const fuseOptions = useMemo(() => ({
    keys: ['name', 'type'],
    threshold: 0.6,
    minMatchCharLength: 1,
    includeScore: true,
    includeMatches: true,
    ignoreLocation: true,
    findAllMatches: true
  }), [searchOptions])
  
  // Perform fuzzy search
  const searchResults = useMemo(() => {
    const trimmedQuery = searchQuery.trim()
    
    // If no search query, return all files
    if (!trimmedQuery) {
      return files
    }
    
    // If query is too short, return all files
    if (trimmedQuery.length < (searchOptions.minMatchCharLength || 1)) {
      return files
    }
    
    // Create a fresh Fuse instance for this search to avoid timing issues
    const fuse = new Fuse(files, fuseOptions)
    
    // Perform fuzzy search
    const results = fuse.search(trimmedQuery)
    
    // Extract the file items from Fuse results
    return results.map(result => result.item)
  }, [searchQuery, files, fuseOptions, searchOptions.minMatchCharLength])
  
  const clearSearch = () => {
    setSearchQuery('')
  }
  
  const isSearching = searchQuery.trim().length >= (searchOptions.minMatchCharLength || 2)
  
  return {
    searchQuery,
    setSearchQuery,
    searchResults,
    clearSearch,
    isSearching
  }
}
