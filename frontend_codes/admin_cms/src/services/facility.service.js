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

export const facilityService = {
  // Get all facilities - uses BOOKING service directly (port 8083)
  async getFacilities(params = {}) {
    const response = await axios.get(`${API_ENDPOINTS.BOOKING_URL}/v1/facilities`, {
      params,
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Get facility by ID - uses BOOKING service directly (port 8083)
  async getFacilityById(id) {
    const response = await axios.get(`${API_ENDPOINTS.BOOKING_URL}/v1/facilities/${id}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Create facility - uses BOOKING service directly (port 8083)
  async createFacility(facilityData) {
    const response = await axios.post(`${API_ENDPOINTS.BOOKING_URL}/v1/facilities`, facilityData, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Update facility - uses BOOKING service directly (port 8083)
  async updateFacility(id, facilityData) {
    const response = await axios.put(`${API_ENDPOINTS.BOOKING_URL}/v1/facilities/${id}`, facilityData, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Delete facility - uses BOOKING service directly (port 8083)
  async deleteFacility(id) {
    const response = await axios.delete(`${API_ENDPOINTS.BOOKING_URL}/v1/facilities/${id}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Get facility schedule - uses BOOKING service directly (port 8083)
  async getFacilitySchedule(id, params = {}) {
    const response = await axios.get(`${API_ENDPOINTS.BOOKING_URL}/v1/facilities/${id}/schedule`, {
      params,
      headers: getAuthHeaders(),
    })
    return response.data
  },
}
