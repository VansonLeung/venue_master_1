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

export const bookingService = {
  // Get all bookings - uses GATEWAY (port 8080) which forwards to booking service
  async getBookings(params = {}) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/bookings`, {
      params,
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Get booking by ID - uses GATEWAY (port 8080) which forwards to booking service
  async getBookingById(id) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/bookings/${id}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Create booking - uses GATEWAY (port 8080) which forwards to booking service
  async createBooking(bookingData) {
    const response = await axios.post(`${API_ENDPOINTS.GATEWAY_URL}/v1/bookings`, bookingData, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Update booking status - uses GATEWAY (port 8080) which forwards to booking service
  async updateBookingStatus(id, status) {
    const response = await axios.patch(`${API_ENDPOINTS.GATEWAY_URL}/v1/bookings/${id}/status`, { status }, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Cancel booking - uses GATEWAY (port 8080) which forwards to booking service
  async cancelBooking(id) {
    const response = await axios.patch(`${API_ENDPOINTS.GATEWAY_URL}/v1/bookings/${id}/cancel`, {}, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Confirm booking - uses GATEWAY (port 8080) which forwards to booking service
  async confirmBooking(id) {
    const response = await axios.post(`${API_ENDPOINTS.GATEWAY_URL}/v1/bookings/${id}/confirm`, {}, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Get booking statistics - uses GATEWAY (port 8080) which forwards to booking service
  async getStats(params = {}) {
    const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/bookings/stats`, {
      params,
      headers: getAuthHeaders(),
    })
    return response.data
  },
}
