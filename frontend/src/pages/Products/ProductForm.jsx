import React, { useState, useEffect, useRef } from 'react'
import api from '../../api'
import toast from 'react-hot-toast'
import { XMarkIcon, PhotoIcon } from '@heroicons/react/24/outline'

const ProductForm = ({ product, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    product_code: '',
    product_name: '',
    category: '',
    unit: '',
    min_stock: 0,
    description: '',
    image: ''
  })
  const [loading, setLoading] = useState(false)
  const [previewImage, setPreviewImage] = useState('')
  const [uploadProgress, setUploadProgress] = useState(0)
  const fileInputRef = useRef(null)

  useEffect(() => {
    if (product) {
      setFormData({
        product_code: product.product_code || '',
        product_name: product.product_name || '',
        category: product.category || '',
        unit: product.unit || '',
        min_stock: product.min_stock || 0,
        description: product.description || '',
        image: product.image || ''
      })
      setPreviewImage(product.image || '')
    }
  }, [product])

  const handleImageChange = (e) => {
    const file = e.target.files[0]
    if (!file) {
      console.log('No file selected')
      return
    }
    
    console.log('📸 File selected:', file.name, file.type, file.size)
    
    // Validate file type
    const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/jpg']
    if (!validTypes.includes(file.type)) {
      toast.error('Please select a valid image (JPG, PNG, GIF, WEBP)')
      return
    }
    
    // Validate file size (max 2MB)
    if (file.size > 2 * 1024 * 1024) {
      toast.error('Image size must be less than 2MB')
      return
    }
    
    setUploadProgress(0)
    
    const reader = new FileReader()
    reader.onload = (event) => {
      const img = new Image()
      img.onload = () => {
        // Resize image to max 600x600 to reduce size
        const maxWidth = 600
        const maxHeight = 600
        let width = img.width
        let height = img.height
        
        if (width > maxWidth || height > maxHeight) {
          const ratio = Math.min(maxWidth / width, maxHeight / height)
          width = Math.round(width * ratio)
          height = Math.round(height * ratio)
        }
        
        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')
        ctx.drawImage(img, 0, 0, width, height)
        
        // Compress to JPEG with 0.7 quality
        const base64String = canvas.toDataURL('image/jpeg', 0.7)
        console.log('✅ Image resized and compressed, length:', base64String.length)
        
        setPreviewImage(base64String)
        setFormData(prev => ({ ...prev, image: base64String }))
        setUploadProgress(100)
        toast.success('Image loaded successfully!')
      }
      img.src = event.target.result
    }
    reader.onerror = (error) => {
      console.error('❌ Error reading file:', error)
      toast.error('Failed to read image file')
    }
    reader.readAsDataURL(file)
  }

  const handleRemoveImage = () => {
    console.log('🗑️ Removing image')
    setPreviewImage('')
    setFormData(prev => ({ ...prev, image: '' }))
    setUploadProgress(0)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)

    try {
      // Debug: Log data sebelum dikirim
      console.log('📤 Submitting product data:')
      console.log('  - Code:', formData.product_code)
      console.log('  - Name:', formData.product_name)
      console.log('  - Image exists:', !!formData.image)
      console.log('  - Image length:', formData.image?.length || 0)
      
      const payload = { ...formData }
      
      if (product) {
        const response = await api.put(`/products/${product.id}`, payload)
        console.log('✅ Update response:', response.data)
        toast.success('Product updated successfully')
      } else {
        const response = await api.post('/products', payload)
        console.log('✅ Create response:', response.data)
        toast.success('Product created successfully')
      }
      
      onSuccess()
      onClose()
    } catch (error) {
      console.error('❌ Error saving product:', error)
      console.error('❌ Error response:', error.response?.data)
      toast.error(error.response?.data?.message || 'Failed to save product')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
      <div className="bg-white rounded-xl shadow-xl max-w-lg w-full max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between p-6 border-b border-gray-200 sticky top-0 bg-white z-10">
          <h3 className="text-lg font-semibold">
            {product ? 'Edit Product' : 'Add New Product'}
          </h3>
          <button
            onClick={onClose}
            className="p-1 rounded-lg hover:bg-gray-100 transition-colors"
            type="button"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Image Upload */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Product Image (Optional)
            </label>
            <div className="flex items-center gap-4">
              {previewImage ? (
                <div className="relative">
                  <img
                    src={previewImage}
                    alt="Preview"
                    className="w-20 h-20 rounded-lg object-cover border border-gray-200"
                  />
                  <button
                    type="button"
                    onClick={handleRemoveImage}
                    className="absolute -top-2 -right-2 p-1 bg-red-500 text-white rounded-full hover:bg-red-600 transition-colors"
                  >
                    <XMarkIcon className="h-4 w-4" />
                  </button>
                </div>
              ) : (
                <div className="w-20 h-20 rounded-lg border-2 border-dashed border-gray-300 flex items-center justify-center bg-gray-50">
                  <PhotoIcon className="h-8 w-8 text-gray-400" />
                </div>
              )}
              <div className="flex-1">
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  className="px-4 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                  disabled={loading}
                >
                  {previewImage ? 'Change Image' : 'Upload Image'}
                </button>
                <p className="text-xs text-gray-500 mt-1">
                  JPG, PNG (max 2MB, akan diresize)
                </p>
                {uploadProgress > 0 && uploadProgress < 100 && (
                  <div className="mt-2 w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                      style={{ width: `${uploadProgress}%` }}
                    />
                  </div>
                )}
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/jpeg,image/png,image/gif,image/webp"
                  onChange={handleImageChange}
                  className="hidden"
                  disabled={loading}
                />
              </div>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Product Code *
            </label>
            <input
              type="text"
              value={formData.product_code}
              onChange={(e) => setFormData({ ...formData, product_code: e.target.value })}
              className="input-field"
              required
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Product Name *
            </label>
            <input
              type="text"
              value={formData.product_name}
              onChange={(e) => setFormData({ ...formData, product_name: e.target.value })}
              className="input-field"
              required
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Category
            </label>
            <input
              type="text"
              value={formData.category}
              onChange={(e) => setFormData({ ...formData, category: e.target.value })}
              className="input-field"
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Unit *
            </label>
            <input
              type="text"
              value={formData.unit}
              onChange={(e) => setFormData({ ...formData, unit: e.target.value })}
              className="input-field"
              required
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Minimum Stock
            </label>
            <input
              type="number"
              value={formData.min_stock}
              onChange={(e) => setFormData({ ...formData, min_stock: parseInt(e.target.value) || 0 })}
              className="input-field"
              min="0"
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="input-field"
              rows="3"
              disabled={loading}
            />
          </div>

          <div className="flex gap-3 pt-4 sticky bottom-0 bg-white py-4 border-t border-gray-100">
            <button
              type="submit"
              className="btn-primary flex-1"
              disabled={loading}
            >
              {loading ? 'Saving...' : (product ? 'Update Product' : 'Create Product')}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="btn-secondary flex-1"
              disabled={loading}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default ProductForm