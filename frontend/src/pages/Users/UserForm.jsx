import React, { useState, useEffect, useRef } from 'react'
import api from '../../api'
import toast from 'react-hot-toast'
import { XMarkIcon, CameraIcon } from '@heroicons/react/24/outline'

const UserForm = ({ user, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    username: '',
    fullname: '',
    password: '',
    role: 'produksi',
    avatar: ''
  })
  const [loading, setLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [previewUrl, setPreviewUrl] = useState(null)
  const fileInputRef = useRef(null)

  useEffect(() => {
    if (user) {
      setFormData({
        username: user.username || '',
        fullname: user.fullname || '',
        password: '',
        role: user.role || 'produksi',
        avatar: user.avatar || user.profile_image || ''
      })
      // Set preview URL jika ada avatar
      if (user.avatar || user.profile_image) {
        const avatarUrl = user.avatar || user.profile_image
        const imageUrl = avatarUrl.startsWith('http') 
          ? avatarUrl 
          : `http://localhost:5000${avatarUrl}`
        setPreviewUrl(imageUrl)
      }
    }
  }, [user])

  const handleImageClick = () => {
    fileInputRef.current?.click()
  }

  const handleFileChange = async (e) => {
    const file = e.target.files[0]
    if (!file) return

    // Validate file size (max 2MB)
    if (file.size > 2 * 1024 * 1024) {
      toast.error('Image size must be less than 2MB')
      e.target.value = ''
      return
    }

    // Validate file type
    if (!file.type.startsWith('image/')) {
      toast.error('Please upload an image file')
      e.target.value = ''
      return
    }

    setUploading(true)
    
    // Preview
    const reader = new FileReader()
    reader.onloadend = () => {
      setPreviewUrl(reader.result)
    }
    reader.readAsDataURL(file)

    try {
      // Upload avatar ke server
      const formDataUpload = new FormData()
      formDataUpload.append('avatar', file)
      
      const response = await api.post('/users/profile/upload-avatar', formDataUpload, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
      
      if (response.data.success) {
        const avatarUrl = response.data.data?.avatar_url || response.data.data?.avatar_url
        setFormData(prev => ({
          ...prev,
          avatar: avatarUrl
        }))
        toast.success('Avatar uploaded successfully!')
      } else {
        toast.error('Failed to upload avatar')
      }
    } catch (error) {
      console.error('Upload error:', error)
      toast.error(error.response?.data?.message || 'Failed to upload avatar')
    } finally {
      setUploading(false)
      e.target.value = ''
    }
  }

  const handleRemoveAvatar = () => {
    setPreviewUrl(null)
    setFormData(prev => ({
      ...prev,
      avatar: ''
    }))
  }

  const getInitials = () => {
    const name = formData.fullname || formData.username || 'U'
    return name.charAt(0).toUpperCase()
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)

    try {
      if (user) {
        await api.put(`/users/${user.id}`, formData)
        toast.success('User updated successfully')
      } else {
        await api.post('/users', formData)
        toast.success('User created successfully')
      }
      onSuccess()
      onClose()
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failed to save user')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
      <div className="bg-white rounded-xl shadow-xl max-w-md w-full">
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h3 className="text-lg font-semibold">
            {user ? 'Edit User' : 'Add New User'}
          </h3>
          <button
            onClick={onClose}
            className="p-1 rounded-lg hover:bg-gray-100 transition-colors"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Avatar Upload */}
          <div className="flex flex-col items-center">
            <div className="relative">
              {previewUrl ? (
                <img
                  src={previewUrl}
                  alt="Avatar"
                  className="h-24 w-24 rounded-full object-cover border-4 border-gray-200"
                />
              ) : (
                <div className="h-24 w-24 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white text-3xl font-semibold border-4 border-gray-200">
                  {getInitials()}
                </div>
              )}
              <button
                type="button"
                onClick={handleImageClick}
                disabled={uploading}
                className="absolute bottom-0 right-0 p-1.5 bg-blue-600 text-white rounded-full hover:bg-blue-700 transition-colors shadow-lg disabled:opacity-50"
              >
                <CameraIcon className="h-5 w-5" />
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleFileChange}
                className="hidden"
                disabled={uploading}
              />
            </div>
            <div className="flex gap-2 mt-2">
              <p className="text-xs text-gray-500">Click camera to upload (max 2MB)</p>
              {previewUrl && (
                <button
                  type="button"
                  onClick={handleRemoveAvatar}
                  className="text-xs text-red-500 hover:text-red-700"
                >
                  Remove
                </button>
              )}
            </div>
            {uploading && (
              <div className="mt-1 text-xs text-blue-600">Uploading...</div>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Username *
            </label>
            <input
              type="text"
              value={formData.username}
              onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              className="input-field"
              required
              disabled={!!user}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Full Name *
            </label>
            <input
              type="text"
              value={formData.fullname}
              onChange={(e) => setFormData({ ...formData, fullname: e.target.value })}
              className="input-field"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {user ? 'New Password (leave blank to keep current)' : 'Password *'}
            </label>
            <input
              type="password"
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              className="input-field"
              required={!user}
              minLength={6}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Role *
            </label>
            <select
              value={formData.role}
              onChange={(e) => setFormData({ ...formData, role: e.target.value })}
              className="input-field"
              required
            >
              <option value="produksi">Produksi</option>
              <option value="head">Head</option>
              <option value="superadmin">Super Admin</option>
            </select>
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="submit"
              className="btn-primary flex-1"
              disabled={loading || uploading}
            >
              {loading ? 'Saving...' : (user ? 'Update' : 'Create')}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="btn-secondary flex-1"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default UserForm