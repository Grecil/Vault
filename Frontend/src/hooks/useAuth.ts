import { useAuth as useClerkAuth, useUser } from '@clerk/clerk-react'
import { useEffect, useState } from 'react'
import { apiClient } from '../lib/api'

export interface UserProfile {
  id: string
  email: string
  role: string
  storageUsed: number
  storageQuota: number
  createdAt: string
}

export const useAuth = () => {
  const { getToken, isSignedIn, isLoaded } = useClerkAuth()
  const { user } = useUser()
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Get or create user profile from backend
  const fetchUserProfile = async () => {
    if (!isSignedIn) return

    setLoading(true)
    setError(null)
    
    try {
      const profile = await apiClient.getUserProfile(getToken)
      setUserProfile(profile)
    } catch (err: any) {
      setError(err.message || 'Failed to load user profile')
      console.error('Error loading user profile:', err)
    } finally {
      setLoading(false)
    }
  }

  // Load user profile when authentication state changes
  useEffect(() => {
    if (isLoaded && isSignedIn) {
      fetchUserProfile()
    } else if (isLoaded && !isSignedIn) {
      setUserProfile(null)
      setError(null)
    }
  }, [isLoaded, isSignedIn])

  return {
    user,
    userProfile,
    isSignedIn,
    isLoaded,
    loading,
    error,
    getToken,
    refreshProfile: fetchUserProfile
  }
}
