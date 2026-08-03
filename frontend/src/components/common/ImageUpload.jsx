import React, { useState, useRef } from 'react'
import { CameraIcon } from '@heroicons/react/24/outline'
import toast from 'react-hot-toast'

const ImageUpload = ({ label, value, onChange, className = '' }) => {
  const [preview, setPreview] = useState(value || null)
  const [uploading, setUploading] = useState(false)
  const fileInputRef = useRef(null)

  const handleImageClick = () => {
    fileInputRef.current?.click()
  }

  const handleFileChange = async (e) => {
    const file = e.target.files[0]
    if (!file) return

    if (file.size > 2 * 1024 * 1024) {
      toast.error('Image size must be less than 2MB')
      e.target.value = ''
      return
    }

    if (!file.type.startsWith('image/')) {
      toast.error('Please upload an image file')
      e.target.value = ''
      return
    }

    setUploading(true)
    
    const reader = new FileReader()
    reader.onloadend = () => {
      setPreview(reader.result)
      onChange(reader.result)
    }
    reader.readAsDataURL(file)

    setUploading(false)
    e.target.value = ''
  }

  const handleRemove = () => {
    setPreview(null)
    onChange('')
  }

  const getInitials = () => {
    return 'U'
  }

  return (
    <div className={`flex flex-col items-center ${className}`}>
      {label && (
        <label className="block text-sm font-medium text-gray-700 mb-2">
          {label}
        </label>
      )}
      <div className="relative">
        {preview ? (
          <img
            src={preview}
            alt="Upload"
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
        {preview && (
          <button
            type="button"
            onClick={handleRemove}
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
  )
}

export default ImageUpload