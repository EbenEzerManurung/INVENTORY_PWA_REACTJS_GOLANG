import React, { createContext, useState, useContext, useEffect } from 'react'
import api from '../api'
import toast from 'react-hot-toast'

const AuthContext = createContext()

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null)
  const [token, setToken] = useState(localStorage.getItem('token'))
  const [loading, setLoading] = useState(true)

  // Load user dari localStorage
  useEffect(() => {
    const loadUser = async () => {
      const storedToken = localStorage.getItem('token')
      const storedUser = localStorage.getItem('user')
      
      if (storedToken && storedUser) {
        try {
          setToken(storedToken)
          setUser(JSON.parse(storedUser))
          api.defaults.headers.common['Authorization'] = `Bearer ${storedToken}`
          
          // Verifikasi token dengan server
          const response = await api.get('/auth/me')
          if (response.data.success) {
            setUser(response.data.data)
            localStorage.setItem('user', JSON.stringify(response.data.data))
          }
        } catch (error) {
          console.error('Failed to load user:', error)
          localStorage.removeItem('token')
          localStorage.removeItem('user')
          delete api.defaults.headers.common['Authorization']
          setToken(null)
          setUser(null)
        }
      }
      setLoading(false)
    }

    loadUser()
  }, [])

  const login = async (username, password) => {
    try {
      console.log('🔄 Login attempt:', username)
      const response = await api.post('/auth/login', { username, password })
      console.log('✅ Login response:', response.data)
      
      // Cek struktur response
      if (!response.data.success) {
        throw new Error(response.data.message || 'Login failed')
      }
      
      const { access_token, refresh_token, user } = response.data.data
      
      console.log('🔑 Token received:', access_token ? access_token.substring(0, 20) + '...' : 'null')
      console.log('👤 User:', user?.fullname || user?.username)
      
      // Simpan ke localStorage
      localStorage.setItem('token', access_token)
      localStorage.setItem('refresh_token', refresh_token)
      localStorage.setItem('user', JSON.stringify(user))
      
      // Set header Authorization
      api.defaults.headers.common['Authorization'] = `Bearer ${access_token}`
      setToken(access_token)
      setUser(user)
      
      toast.success(`Welcome back, ${user.fullname || user.username}!`)
      return { success: true, user }
    } catch (error) {
      console.error('❌ Login error:', error.response?.data || error.message)
      const errorMsg = error.response?.data?.message || error.message || 'Login failed'
      toast.error(errorMsg)
      return { success: false, error: errorMsg }
    }
  }

  const logout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
    delete api.defaults.headers.common['Authorization']
    setToken(null)
    setUser(null)
    toast.success('Logged out successfully')
  }

  const updateProfile = async (data) => {
    try {
      console.log('🔄 Updating profile...', data)
      const response = await api.put('/users/profile', data)
      console.log('✅ Update profile response:', response.data)
      
      if (response.data.success) {
        setUser(response.data.data)
        localStorage.setItem('user', JSON.stringify(response.data.data))
        toast.success('Profile updated successfully!')
        return { success: true }
      }
      return { success: false, error: response.data.message || 'Failed to update profile' }
    } catch (error) {
      console.error('❌ Update profile error:', error.response?.data)
      toast.error(error.response?.data?.message || 'Failed to update profile')
      return { success: false, error: error.response?.data?.message }
    }
  }

  const changePassword = async (currentPassword, newPassword) => {
    try {
      console.log('🔄 Changing password...')
      const response = await api.put('/users/profile/password', {
        currentPassword,
        newPassword
      })
      console.log('✅ Change password response:', response.data)
      
      if (response.data.success) {
        toast.success('Password changed successfully!')
        return { success: true }
      }
      return { success: false, error: response.data.message || 'Failed to change password' }
    } catch (error) {
      console.error('❌ Change password error:', error.response?.data)
      toast.error(error.response?.data?.message || 'Failed to change password')
      return { success: false, error: error.response?.data?.message }
    }
  }

  const uploadAvatar = async (file) => {
    try {
      console.log('🔄 Uploading avatar...')
      console.log('📤 File:', file.name, file.size, file.type)
      
      const formData = new FormData()
      formData.append('avatar', file)
      
      // Log FormData entries
      for (let pair of formData.entries()) {
        console.log('📤 FormData:', pair[0], pair[1])
      }
      
      const response = await api.post('/users/profile/upload-avatar', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
      
      console.log('✅ Upload response:', response.data)
      
      if (response.data.success) {
        // Ambil data user terbaru
        const userResponse = await api.get('/auth/me')
        console.log('✅ User response:', userResponse.data)
        
        if (userResponse.data.success) {
          const updatedUser = userResponse.data.data
          setUser(updatedUser)
          localStorage.setItem('user', JSON.stringify(updatedUser))
          console.log('✅ User updated:', updatedUser)
        }
        toast.success('Avatar uploaded successfully!')
        return { success: true, data: response.data.data }
      }
      return { success: false, error: response.data.message || 'Failed to upload avatar' }
    } catch (error) {
      console.error('❌ Upload avatar error:', error)
      console.error('❌ Error response:', error.response?.data)
      toast.error(error.response?.data?.message || 'Failed to upload avatar')
      return { success: false, error: error.response?.data?.message }
    }
  }

  const removeAvatar = async () => {
    try {
      console.log('🔄 Removing avatar...')
      const response = await api.delete('/users/profile/avatar')
      console.log('✅ Remove avatar response:', response.data)
      
      if (response.data.success) {
        // Ambil data user terbaru
        const userResponse = await api.get('/auth/me')
        if (userResponse.data.success) {
          const updatedUser = userResponse.data.data
          setUser(updatedUser)
          localStorage.setItem('user', JSON.stringify(updatedUser))
        }
        toast.success('Avatar removed successfully!')
        return { success: true }
      }
      return { success: false, error: response.data.message || 'Failed to remove avatar' }
    } catch (error) {
      console.error('❌ Remove avatar error:', error.response?.data)
      toast.error(error.response?.data?.message || 'Failed to remove avatar')
      return { success: false, error: error.response?.data?.message }
    }
  }

  const hasPermission = (requiredRoles) => {
    if (!user) return false
    if (!requiredRoles || requiredRoles.length === 0) return true
    return requiredRoles.includes(user.role)
  }

  const value = {
    user,
    setUser,
    token,
    loading,
    login,
    logout,
    updateProfile,
    changePassword,
    uploadAvatar,
    removeAvatar,
    hasPermission,
    isAuthenticated: !!token && !!user
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}