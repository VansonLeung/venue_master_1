import { createContext, useContext, useState, useEffect } from 'react'
import { authService } from '@/services/auth.service'

const AuthContext = createContext(null)

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Check if user is already logged in
    const storedUser = authService.getCurrentUser()
    if (storedUser && authService.isAuthenticated()) {
      setUser(storedUser)
    }
    setLoading(false)
  }, [])

  const login = async (email, password) => {
    const data = await authService.login(email, password)
    setUser(data.user)
    return data
  }

  const register = async (userData) => {
    const data = await authService.register(userData)
    setUser(data.user)
    return data
  }

  const logout = () => {
    authService.logout()
    setUser(null)
  }

  const value = {
    user,
    login,
    register,
    logout,
    isAuthenticated: authService.isAuthenticated(),
    isAdmin: authService.isAdmin(),
    loading,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
