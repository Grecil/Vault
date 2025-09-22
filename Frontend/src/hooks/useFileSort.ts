import { useState, useMemo } from 'react'
import { type FileItem } from '../components/FileGrid'

type SortField = 'name' | 'size' | 'type' | 'uploadDate' | 'visibility'
type SortOrder = 'asc' | 'desc' | 'none'

interface UseFileSortReturn {
  sortField: SortField | null
  sortOrder: SortOrder
  sortedFiles: FileItem[]
  handleSort: (field: SortField) => void
}

// Helper function to parse size strings like "1.2 MB" back to bytes for proper sorting
const parseSize = (sizeStr: string): number => {
  const match = sizeStr.match(/^([\d.]+)\s*(B|KB|MB|GB)$/i)
  if (!match) return 0
  
  const [, value, unit] = match
  const multipliers: Record<string, number> = { 
    B: 1, 
    KB: 1024, 
    MB: 1024**2, 
    GB: 1024**3 
  }
  return parseFloat(value) * (multipliers[unit.toUpperCase()] || 0)
}

// Helper function to parse date strings for sorting
const parseDate = (dateStr: string): number => {
  if (dateStr === 'Unknown') return 0
  const date = new Date(dateStr)
  return isNaN(date.getTime()) ? 0 : date.getTime()
}

export const useFileSort = (files: FileItem[]): UseFileSortReturn => {
  const [sortField, setSortField] = useState<SortField | null>(null)
  const [sortOrder, setSortOrder] = useState<SortOrder>('none')

  // Handle sort cycling: none -> asc -> desc -> none
  const handleSort = (field: SortField) => {
    if (sortField !== field) {
      // New field, start with ascending
      setSortField(field)
      setSortOrder('asc')
    } else {
      // Same field, cycle through orders
      if (sortOrder === 'none') {
        setSortOrder('asc')
      } else if (sortOrder === 'asc') {
        setSortOrder('desc')
      } else {
        setSortOrder('none')
        setSortField(null)
      }
    }
  }

  // Sort files based on current sort state
  const sortedFiles = useMemo(() => {
    if (!sortField || sortOrder === 'none') {
      return files
    }

    const sorted = [...files].sort((a, b) => {
      let comparison = 0

      switch (sortField) {
        case 'name':
          comparison = a.name.localeCompare(b.name)
          break
        case 'size':
          comparison = parseSize(a.size) - parseSize(b.size)
          break
        case 'type':
          comparison = a.type.localeCompare(b.type)
          break
        case 'uploadDate':
          comparison = parseDate(a.uploadDate) - parseDate(b.uploadDate)
          break
        case 'visibility':
          // Public files first, then private
          comparison = Number(b.isPublic) - Number(a.isPublic)
          break
        default:
          return 0
      }

      return sortOrder === 'desc' ? -comparison : comparison
    })

    return sorted
  }, [files, sortField, sortOrder])

  return {
    sortField,
    sortOrder,
    sortedFiles,
    handleSort
  }
}
