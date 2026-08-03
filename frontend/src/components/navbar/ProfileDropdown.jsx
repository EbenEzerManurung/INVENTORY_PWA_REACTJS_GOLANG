import React, { useState, useRef, useEffect } from 'react'
import { useAuth } from '../../context/AuthContext'
import {
  UserCircleIcon,
  KeyIcon,
  ArrowRightOnRectangleIcon,
  CameraIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline'
import toast from 'react-hot-toast'

const ProfileDropdown = () => {
  const { user, setUser, logout, updateProfile, changePassword, uploadAvatar, removeAvatar } = useAuth()
  const [isOpen, setIsOpen] = useState(false)
  const [isEditingProfile, setIsEditingProfile] = useState(false)
  const [isEditingPassword, setIsEditingPassword] = useState(false)
  const dropdownRef = useRef(null)

  const [formData, setFormData] = useState({
    fullname: user?.fullname || '',
    username: user?.username || '',
    email: user?.email || ''
  })
  const [passwordData, setPasswordData] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  })
  const [selectedFile, setSelectedFile] = useState(null)
  const [previewUrl, setPreviewUrl] = useState(null)
  const [loading, setLoading] = useState(false)
  const fileInputRef = useRef(null)

  // ✅ Set preview URL from user
  useEffect(() => {
    if (user?.profile_image) {
      // Pastikan URL gambar lengkap
      const imageUrl = user.profile_image.startsWith('http') 
        ? user.profile_image 
        : `http://localhost:5000${user.profile_image}`
      setPreviewUrl(imageUrl)
      console.log('🖼️ Preview URL set:', imageUrl)
    } else if (user?.avatar) {
      const imageUrl = user.avatar.startsWith('http') 
        ? user.avatar 
        : `http://localhost:5000${user.avatar}`
      setPreviewUrl(imageUrl)
    } else {
      setPreviewUrl(null)
    }
  }, [user])

  // Close dropdown on click outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setIsOpen(false)
        setIsEditingProfile(false)
        setIsEditingPassword(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleImageClick = () => {
    fileInputRef.current?.click()
  }

  const handleFileChange = async (e) => {
    const file = e.target.files[0]
    if (!file) return

    console.log('🔄 File selected:', file.name, file.size, file.type)

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

    setSelectedFile(file)
    setLoading(true)
    
    // Preview
    const reader = new FileReader()
    reader.onloadend = () => {
      setPreviewUrl(reader.result)
    }
    reader.readAsDataURL(file)

    try {
      const result = await uploadAvatar(file)
      console.log('✅ Upload result:', result)
      
      if (result.success) {
        setSelectedFile(null)
        // ✅ Ambil data user terbaru dari response
        if (result.data?.user) {
          const updatedUser = result.data.user
          setUser(updatedUser)
          localStorage.setItem('user', JSON.stringify(updatedUser))
          // ✅ Set preview URL dari user yang diupdate
          if (updatedUser.profile_image) {
            const imageUrl = updatedUser.profile_image.startsWith('http') 
              ? updatedUser.profile_image 
              : `http://localhost:5000${updatedUser.profile_image}`
            setPreviewUrl(imageUrl)
          }
          console.log('✅ User updated with avatar:', updatedUser)
        }
        toast.success('Avatar uploaded successfully!')
      } else {
        // Reset preview jika gagal
        setPreviewUrl(user?.profile_image ? `http://localhost:5000${user.profile_image}` : null)
        toast.error(result.error || 'Failed to upload avatar')
      }
    } catch (error) {
      console.error('❌ Upload error:', error)
      setPreviewUrl(user?.profile_image ? `http://localhost:5000${user.profile_image}` : null)
      toast.error('Failed to upload avatar')
    } finally {
      setLoading(false)
      e.target.value = ''
    }
  }

  const handleUpdateProfile = async (e) => {
    e.preventDefault()
    setLoading(true)

    const result = await updateProfile({
      fullname: formData.fullname,
      username: formData.username,
      email: formData.email
    })

    setLoading(false)
    if (result.success) {
      setIsEditingProfile(false)
    }
  }

  const handleUpdatePassword = async (e) => {
    e.preventDefault()

    if (passwordData.newPassword !== passwordData.confirmPassword) {
      toast.error('New password and confirm password do not match')
      return
    }

    if (passwordData.newPassword.length < 6) {
      toast.error('New password must be at least 6 characters')
      return
    }

    setLoading(true)
    const result = await changePassword(passwordData.currentPassword, passwordData.newPassword)
    setLoading(false)

    if (result.success) {
      setIsEditingPassword(false)
      setPasswordData({
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
      })
    }
  }

  const handleRemoveAvatar = async () => {
    if (!window.confirm('Are you sure you want to remove your profile picture?')) return
    setLoading(true)
    const result = await removeAvatar()
    setLoading(false)
    if (result.success) {
      setPreviewUrl(null)
    }
  }

  const handleLogout = async () => {
    setIsOpen(false)
    await logout()
  }

  const getInitials = () => {
    if (user?.fullname) {
      return user.fullname.charAt(0).toUpperCase()
    }
    if (user?.username) {
      return user.username.charAt(0).toUpperCase()
    }
    return 'U'
  }

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Profile Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 p-1 rounded-full hover:bg-gray-100 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        {previewUrl ? (
          <img
            src={previewUrl}
            alt="Profile"
            className="h-10 w-10 rounded-full object-cover border-2 border-gray-200"
            onError={() => {
              console.log('❌ Image load error, using fallback')
              setPreviewUrl(null)
            }}
          />
        ) : (
          <div className="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-semibold text-lg">
            {getInitials()}
          </div>
        )}
        <span className="hidden md:block text-sm font-medium text-gray-700">
          {user?.fullname || user?.username}
        </span>
        <svg
          className={`hidden md:block h-4 w-4 text-gray-500 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 bg-white rounded-xl shadow-lg border border-gray-200 overflow-hidden z-50">
          {/* User Info */}
          <div className="px-4 py-3 bg-gradient-to-r from-blue-50 to-indigo-50 border-b border-gray-200">
            <div className="flex items-center gap-3">
              {previewUrl ? (
                <img
                  src={previewUrl}
                  alt="Profile"
                  className="h-12 w-12 rounded-full object-cover border-2 border-white shadow-md"
                  onError={() => setPreviewUrl(null)}
                />
              ) : (
                <div className="h-12 w-12 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-semibold text-xl shadow-md">
                  {getInitials()}
                </div>
              )}
              <div className="flex-1 min-w-0">
                <p className="text-sm font-semibold text-gray-900 truncate">
                  {user?.fullname || user?.username}
                </p>
                <p className="text-xs text-gray-600 truncate">@{user?.username}</p>
                <p className="text-xs text-gray-500 capitalize">{user?.role}</p>
              </div>
            </div>
          </div>

          {/* Menu Items */}
          <div className="py-1">
            <button
              onClick={() => {
                setIsEditingProfile(true)
                setIsEditingPassword(false)
                setIsOpen(false)
              }}
              className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-3 transition-colors"
            >
              <UserCircleIcon className="h-5 w-5 text-gray-500" />
              Edit Profile
            </button>

            <button
              onClick={() => {
                setIsEditingPassword(true)
                setIsEditingProfile(false)
                setIsOpen(false)
              }}
              className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-3 transition-colors"
            >
              <KeyIcon className="h-5 w-5 text-gray-500" />
              Change Password
            </button>

            <hr className="my-1 border-gray-200" />

            <button
              onClick={handleLogout}
              className="w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50 flex items-center gap-3 transition-colors"
            >
              <ArrowRightOnRectangleIcon className="h-5 w-5" />
              Logout
            </button>
          </div>
        </div>
      )}

      {/* Edit Profile Modal */}
      {isEditingProfile && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-2xl max-w-md w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-semibold text-gray-900">Edit Profile</h2>
                <button
                  onClick={() => {
                    setIsEditingProfile(false)
                    setSelectedFile(null)
                  }}
                  className="p-1 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <XMarkIcon className="h-6 w-6 text-gray-500" />
                </button>
              </div>

              <form onSubmit={handleUpdateProfile} className="space-y-4">
                {/* Profile Image Upload */}
                <div className="flex flex-col items-center">
                  <div className="relative">
                    {previewUrl ? (
                      <img
                        src={previewUrl}
                        alt="Profile"
                        className="h-24 w-24 rounded-full object-cover border-4 border-gray-200"
                        onError={() => setPreviewUrl(null)}
                      />
                    ) : (
                      <div className="h-24 w-24 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white text-3xl font-semibold border-4 border-gray-200">
                        {getInitials()}
                      </div>
                    )}
                    <button
                      type="button"
                      onClick={handleImageClick}
                      disabled={loading}
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
                      disabled={loading}
                    />
                  </div>
                  <div className="flex gap-2 mt-2">
                    <p className="text-xs text-gray-500">Click camera to upload (max 2MB)</p>
                    {previewUrl && (user?.profile_image || user?.avatar) && (
                      <button
                        type="button"
                        onClick={handleRemoveAvatar}
                        disabled={loading}
                        className="text-xs text-red-500 hover:text-red-700 disabled:opacity-50"
                      >
                        Remove
                      </button>
                    )}
                  </div>
                </div>

                {/* Username */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Username
                  </label>
                  <input
                    type="text"
                    value={formData.username}
                    onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>

                {/* Fullname */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Full Name
                  </label>
                  <input
                    type="text"
                    value={formData.fullname}
                    onChange={(e) => setFormData({ ...formData, fullname: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>

                {/* Email */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Email
                  </label>
                  <input
                    type="email"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {loading ? (
                    <>
                      <svg className="animate-spin h-5 w-5 text-white" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      Saving...
                    </>
                  ) : (
                    'Save Changes'
                  )}
                </button>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Change Password Modal */}
      {isEditingPassword && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-2xl max-w-md w-full">
            <div className="p-6">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-semibold text-gray-900">Change Password</h2>
                <button
                  onClick={() => {
                    setIsEditingPassword(false)
                    setPasswordData({
                      currentPassword: '',
                      newPassword: '',
                      confirmPassword: '',
                    })
                  }}
                  className="p-1 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <XMarkIcon className="h-6 w-6 text-gray-500" />
                </button>
              </div>

              <form onSubmit={handleUpdatePassword} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Current Password
                  </label>
                  <input
                    type="password"
                    value={passwordData.currentPassword}
                    onChange={(e) => setPasswordData({ ...passwordData, currentPassword: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    New Password
                  </label>
                  <input
                    type="password"
                    value={passwordData.newPassword}
                    onChange={(e) => setPasswordData({ ...passwordData, newPassword: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                    minLength={6}
                  />
                  <p className="text-xs text-gray-500 mt-1">Minimum 6 characters</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Confirm New Password
                  </label>
                  <input
                    type="password"
                    value={passwordData.confirmPassword}
                    onChange={(e) => setPasswordData({ ...passwordData, confirmPassword: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {loading ? (
                    <>
                      <svg className="animate-spin h-5 w-5 text-white" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      Updating...
                    </>
                  ) : (
                    'Update Password'
                  )}
                </button>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default ProfileDropdown