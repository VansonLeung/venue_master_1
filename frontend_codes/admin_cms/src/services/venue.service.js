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

export const venueService = {
  // Get all venues - uses Gateway (port 8080)
  async getVenues(params = {}) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/venues`, {
      params,
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Get venue by ID - uses Gateway (port 8080)
  async getVenueById(id) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/venues/${id}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Create venue - uses Gateway (port 8080)
  async createVenue(venueData) {
    const response = await axios.post(`${API_ENDPOINTS.GATEWAY_URL}/v1/venues`, venueData, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Update venue - uses Gateway (port 8080)
  async updateVenue(id, venueData) {
    const response = await axios.put(`${API_ENDPOINTS.GATEWAY_URL}/v1/venues/${id}`, venueData, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Delete venue - uses Gateway (port 8080)
  async deleteVenue(id) {
    const response = await axios.delete(`${API_ENDPOINTS.GATEWAY_URL}/v1/venues/${id}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },
}
