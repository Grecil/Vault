import { useEffect, useCallback } from 'react'

// Simple event emitter for storage stats refresh with debouncing
class StorageStatsEventEmitter {
  private listeners: (() => void)[] = []
  private debounceTimer: number | null = null
  private readonly debounceMs = 1000 // 1 second debounce

  subscribe(listener: () => void) {
    this.listeners.push(listener)
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener)
    }
  }

  emit() {
    // Clear any existing timer
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer)
    }

    // Set a new timer to emit after debounce period
    this.debounceTimer = window.setTimeout(() => {
      this.listeners.forEach(listener => listener())
      this.debounceTimer = null
    }, this.debounceMs)
  }

  // For immediate refresh (when user clicks refresh button)
  emitImmediate() {
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer)
      this.debounceTimer = null
    }
    this.listeners.forEach(listener => listener())
  }
}

// Global instance
const storageStatsEmitter = new StorageStatsEventEmitter()

// Hook for components that want to trigger storage stats refresh
export const useStorageStatsRefreshTrigger = () => {
  const triggerRefresh = useCallback(() => {
    storageStatsEmitter.emit()
  }, [])

  const triggerRefreshImmediate = useCallback(() => {
    storageStatsEmitter.emitImmediate()
  }, [])

  return { triggerRefresh, triggerRefreshImmediate }
}

// Hook for components that want to listen for storage stats refresh
export const useStorageStatsRefreshListener = (onRefresh: () => void) => {
  useEffect(() => {
    const unsubscribe = storageStatsEmitter.subscribe(onRefresh)
    return unsubscribe
  }, [onRefresh])
}
