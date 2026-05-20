import { useCallback, useEffect, useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { apiFetch } from '../utils/apiFetch'

export function useCurrentUser() {
  const { isAuthenticated, isReady } = useAuth()

  const [currentUser, setCurrentUser] = useState(null)
  const [isLoadingUser, setIsLoadingUser] = useState(false)
  const [userError, setUserError] = useState('')

  const loadCurrentUser = useCallback(async () => {
    if (!isAuthenticated) return

    setIsLoadingUser(true)
    setUserError('')

    try {
      const res = await apiFetch('/api/users/me')
      const data = await res.json().catch(() => null)

      if (!res.ok) {
        throw new Error(data?.error || 'Failed to load current user')
      }

      setCurrentUser(data)
    } catch (error) {
      setUserError(error.message || 'Failed to load current user')
      setCurrentUser(null)
    } finally {
      setIsLoadingUser(false)
    }
  }, [isAuthenticated])

  useEffect(() => {
    if (!isReady || !isAuthenticated) return
    loadCurrentUser()
  }, [isReady, isAuthenticated, loadCurrentUser])

  return {
    currentUser,
    setCurrentUser,
    isLoadingUser,
    userError,
    loadCurrentUser,
  }
}