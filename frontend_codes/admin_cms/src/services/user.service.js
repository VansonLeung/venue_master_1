import axios from 'axios'
import { API_ENDPOINTS } from './api'

// Helper to get auth headers
const getAuthHeaders = () => {
  const token = localStorage.getItem('admin_token')
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  }
}

export const userService = {
  // Get all users - uses Gateway (port 8080)
  async getUsers(params = {}) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/users`, {
      params,
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Get user by ID - uses Gateway (port 8080)
  async getUserById(id) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/users/${id}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Update user - uses Gateway (port 8080)
  async updateUser(id, userData) {
    const response = await axios.put(`${API_ENDPOINTS.GATEWAY_URL}/v1/users/${id}`, userData, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Update user roles - uses Gateway (port 8080)
  async updateUserRoles(id, roles) {
    const response = await axios.patch(`${API_ENDPOINTS.GATEWAY_URL}/v1/users/${id}/roles`, { roles }, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Deactivate user - uses Gateway (port 8080)
  async deactivateUser(id) {
    const response = await axios.patch(`${API_ENDPOINTS.GATEWAY_URL}/v1/users/${id}/deactivate`, {}, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Activate user - uses Gateway (port 8080)
  async activateUser(id) {
    const response = await axios.patch(`${API_ENDPOINTS.GATEWAY_URL}/v1/users/${id}/activate`, {}, {
      headers: getAuthHeaders(),
    })
    return response.data
  },
}
