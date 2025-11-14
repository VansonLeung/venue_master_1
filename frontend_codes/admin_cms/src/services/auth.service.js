import axios from 'axios'
import { API_ENDPOINTS } from './api'

export const authService = {
  // Login - uses AUTH service directly (port 8081)
  async login(email, password) {
    const response = await axios.post(`${API_ENDPOINTS.AUTH_URL}/v1/auth/login`, {
      email,
      password,
    })

    const { accessToken, refreshToken, user } = response.data

    // Store tokens and user info
    localStorage.setItem('admin_token', accessToken)
    localStorage.setItem('admin_refresh_token', refreshToken)
    localStorage.setItem('admin_user', JSON.stringify(user))

    return response.data
  },

  // Register - uses AUTH service directly (port 8081)
  async register(userData) {
    const response = await axios.post(`${API_ENDPOINTS.AUTH_URL}/v1/auth/register`, userData)

    const { accessToken, refreshToken, user } = response.data

    // Store tokens and user info
    localStorage.setItem('admin_token', accessToken)
    localStorage.setItem('admin_refresh_token', refreshToken)
    localStorage.setItem('admin_user', JSON.stringify(user))

    return response.data
  },

  // Logout
  logout() {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_refresh_token')
    localStorage.removeItem('admin_user')
  },

  // Get current user
  getCurrentUser() {
    const userStr = localStorage.getItem('admin_user')
    return userStr ? JSON.parse(userStr) : null
  },

  // Check if user is authenticated
  isAuthenticated() {
    return !!localStorage.getItem('admin_token')
  },

  // Check if user is admin
  isAdmin() {
    const user = this.getCurrentUser()
    return user?.roles?.includes('ADMIN') || user?.roles?.includes('SUPER_ADMIN')
  },
}
