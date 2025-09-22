import { useState, useEffect, useCallback } from 'react'
import { useAuth } from '@clerk/clerk-react'
import { useStorageStatsRefreshListener } from './useStorageStatsRefresh'

export interface StorageStatistics {
  total_storage: number     // Deduplicated storage used in bytes
  original_storage: number  // Storage without deduplication in bytes
  storage_quota: number     // User's storage quota in bytes
  file_count: number        // Total number of files owned
  duplicate_count: number   // Number of duplicate files avoided
  savings: {
    bytes: number          // Bytes saved through deduplication
    percentage: number     // Percentage saved (0-100)
  }
}

interface UseStorageStatsReturn {
  statistics: StorageStatistics | null
  loading: boolean
  error: string | null
  refresh: () => Promise<void>
  formatBytes: (bytes: number) => string
}

export const useStorageStats = (): UseStorageStatsReturn => {
  const [statistics, setStatistics] = useState<StorageStatistics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const { getToken } = useAuth()

  const formatBytes = useCallback((bytes: number): string => {
    if (bytes === 0) return '0 Bytes'
    
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }, [])

  const fetchStatistics = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)

      const token = await getToken()
      if (!token) {
        throw new Error('No authentication token available')
      }

      const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'
      const response = await fetch(`${API_BASE_URL}/user/storage/statistics`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()

      if (data) {
        // Validate the response data structure
        const stats = data as StorageStatistics
        
        // Ensure required fields exist and have valid values
        if (typeof stats.total_storage !== 'number' || stats.total_storage < 0) {
          stats.total_storage = 0
        }
        if (typeof stats.original_storage !== 'number' || stats.original_storage < 0) {
          stats.original_storage = 0
        }
        if (typeof stats.storage_quota !== 'number' || stats.storage_quota < 0) {
          stats.storage_quota = 0
        }
        if (typeof stats.file_count !== 'number' || stats.file_count < 0) {
          stats.file_count = 0
        }
        if (typeof stats.duplicate_count !== 'number' || stats.duplicate_count < 0) {
          stats.duplicate_count = 0
        }
        
        // Validate savings object
        if (!stats.savings || typeof stats.savings !== 'object') {
          stats.savings = { bytes: 0, percentage: 0 }
        }
        if (typeof stats.savings.bytes !== 'number' || stats.savings.bytes < 0) {
          stats.savings.bytes = 0
        }
        if (typeof stats.savings.percentage !== 'number' || stats.savings.percentage < 0 || stats.savings.percentage > 100) {
          stats.savings.percentage = 0
        }
        
        // Ensure logical consistency
        if (stats.original_storage < stats.total_storage) {
          stats.original_storage = stats.total_storage
          stats.savings.bytes = 0
          stats.savings.percentage = 0
        }
        
        setStatistics(stats)
      } else {
        throw new Error('Invalid response format')
      }
    } catch (err: any) {
      console.error('Failed to fetch storage statistics:', err)
      
      // Provide more specific error messages
      if (err.response?.status === 401) {
        setError('Authentication required')
      } else if (err.response?.status === 403) {
        setError('Access denied')
      } else if (err.response?.status === 404) {
        setError('Statistics endpoint not found')
      } else if (err.response?.status >= 500) {
        setError('Server error occurred')
      } else if (err.message) {
        setError(err.message)
      } else {
        setError('Failed to load storage statistics')
      }
      
      // Set default statistics on error to prevent UI crashes
      setStatistics({
        total_storage: 0,
        original_storage: 0,
        storage_quota: 0,
        file_count: 0,
        duplicate_count: 0,
        savings: {
          bytes: 0,
          percentage: 0
        }
      })
    } finally {
      setLoading(false)
    }
  }, [getToken])

  const refresh = useCallback(async () => {
    await fetchStatistics()
  }, [fetchStatistics])

  useEffect(() => {
    fetchStatistics()
  }, [fetchStatistics])

  // Listen for refresh events from other components
  useStorageStatsRefreshListener(refresh)

  return {
    statistics,
    loading,
    error,
    refresh,
    formatBytes
  }
}
