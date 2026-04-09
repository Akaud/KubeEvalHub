import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [accessToken, setAccessToken] = useState(null)
  const [refreshToken, setRefreshToken] = useState(null)
  const [isReady, setIsReady] = useState(false)

  const refreshPromiseRef = useRef(null)

  useEffect(() => {
    const storedAccessToken = localStorage.getItem('accessToken')
    const storedRefreshToken = localStorage.getItem('refreshToken')

    if (storedAccessToken) {
      setAccessToken(storedAccessToken)
    }

    if (storedRefreshToken) {
      setRefreshToken(storedRefreshToken)
    }

    setIsReady(true)
  }, [])

  const login = useCallback((nextAccessToken, nextRefreshToken) => {
    localStorage.setItem('accessToken', nextAccessToken)
    localStorage.setItem('refreshToken', nextRefreshToken)
    setAccessToken(nextAccessToken)
    setRefreshToken(nextRefreshToken)
  }, [])

  const clearAuth = useCallback(() => {
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
    setAccessToken(null)
    setRefreshToken(null)
  }, [])

  const logout = useCallback(async () => {
    const currentRefreshToken = localStorage.getItem('refreshToken')

    try {
      if (currentRefreshToken) {
        await fetch('/api/auth/logout', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refreshToken: currentRefreshToken }),
        })
      }
    } catch {
    } finally {
      clearAuth()
    }
  }, [clearAuth])

  const refreshAccessToken = useCallback(async () => {
    if (refreshPromiseRef.current) {
      return refreshPromiseRef.current
    }

    refreshPromiseRef.current = (async () => {
      const currentRefreshToken = localStorage.getItem('refreshToken')

      if (!currentRefreshToken) {
        clearAuth()
        return null
      }

      try {
        const res = await fetch('/api/auth/refresh', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refreshToken: currentRefreshToken }),
        })

        const data = await res.json().catch(() => null)

        if (!res.ok || !data?.accessToken || !data?.refreshToken) {
          clearAuth()
          return null
        }

        login(data.accessToken, data.refreshToken)
        return data.accessToken
      } catch {
        clearAuth()
        return null
      } finally {
        refreshPromiseRef.current = null
      }
    })()

    return refreshPromiseRef.current
  }, [clearAuth, login])

  const value = useMemo(
    () => ({
      accessToken,
      refreshToken,
      token: accessToken,
      isAuthenticated: Boolean(accessToken),
      isReady,
      login,
      logout,
      clearAuth,
      refreshAccessToken,
    }),
    [accessToken, refreshToken, isReady, login, logout, clearAuth, refreshAccessToken]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)

  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }

  return context
}