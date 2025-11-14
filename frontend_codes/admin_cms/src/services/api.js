import axios from 'axios'

// API Configuration - following the same structure as test-api.sh
const BASE_URL = import.meta.env.VITE_BASE_URL || 'http://localhost'
const GATEWAY_PORT = import.meta.env.VITE_GATEWAY_PORT || '8080'
const AUTH_PORT = import.meta.env.VITE_AUTH_PORT || '8081'
const BOOKING_PORT = import.meta.env.VITE_BOOKING_PORT || '8083'

export const API_ENDPOINTS = {
  GATEWAY_URL: `${BASE_URL}:${GATEWAY_PORT}`,
  AUTH_URL: `${BASE_URL}:${AUTH_PORT}`,
  BOOKING_URL: `${BASE_URL}:${BOOKING_PORT}`,
}

// Default axios instance (uses Gateway for most operations)
const api = axios.create({
  baseURL: API_ENDPOINTS.GATEWAY_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor - add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('admin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor - handle errors
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    // Handle 401 Unauthorized - try to refresh token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        const refreshToken = localStorage.getItem('admin_refresh_token')
        if (refreshToken) {
          // Use AUTH_URL for refresh endpoint
          const response = await axios.post(`${API_ENDPOINTS.AUTH_URL}/v1/auth/refresh`, {
            refreshToken,
          })

          const { accessToken } = response.data
          localStorage.setItem('admin_token', accessToken)

          originalRequest.headers.Authorization = `Bearer ${accessToken}`
          return api(originalRequest)
        }
      } catch (refreshError) {
        // Refresh failed - clear auth and redirect to login
        localStorage.removeItem('admin_token')
        localStorage.removeItem('admin_refresh_token')
        localStorage.removeItem('admin_user')
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

export default api
