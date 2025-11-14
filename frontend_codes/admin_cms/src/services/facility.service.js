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
  // Get all facilities - uses GATEWAY (port 8080) which forwards to booking service
  async getFacilities(params = {}) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/facilities`, {
      params,
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Get facility by ID - uses GATEWAY (port 8080) which forwards to booking service
  async getFacilityById(id) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/facilities/${id}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Create facility - uses GATEWAY (port 8080) which forwards to booking service
  async createFacility(facilityData) {
    const response = await axios.post(`${API_ENDPOINTS.GATEWAY_URL}/v1/facilities`, facilityData, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Update facility - uses GATEWAY (port 8080) which forwards to booking service
  async updateFacility(id, facilityData) {
    const response = await axios.put(`${API_ENDPOINTS.GATEWAY_URL}/v1/facilities/${id}`, facilityData, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Delete facility - uses GATEWAY (port 8080) which forwards to booking service
  async deleteFacility(id) {
    const response = await axios.delete(`${API_ENDPOINTS.GATEWAY_URL}/v1/facilities/${id}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Get facility schedule - uses GATEWAY (port 8080) which forwards to booking service
  async getFacilitySchedule(id, params = {}) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/facilities/${id}/schedule`, {
      params,
      headers: getAuthHeaders(),
    })
    return response.data
  },
}
