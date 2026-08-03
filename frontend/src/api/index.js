import axios from 'axios'

// Gunakan proxy dari Vite
const API_URL = '/api'

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Accept': 'application/json'
  },
  timeout: 30000
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    
    console.log(`📤 ${config.method.toUpperCase()} ${config.url}`)
    console.log('📤 Headers:', config.headers)
    
    if (config.data instanceof FormData) {
      console.log('📤 FormData detected, size:', config.data.get('avatar')?.size)
    }
    
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => {
    console.log(`📥 ${response.status} ${response.config.url}`)
    return response
  },
  (error) => {
    console.error('❌ API Error:', error.response?.status, error.response?.data)
    
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      delete api.defaults.headers.common['Authorization']
      window.location.href = '/login'
    }
    
    if (!error.response) {
      console.error('🔌 Network error - Pastikan backend berjalan di port 5000')
    }
    
    return Promise.reject(error)
  }
)

export default api