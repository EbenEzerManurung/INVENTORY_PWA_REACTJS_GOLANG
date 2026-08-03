import React, { useState, useEffect } from 'react'
import api from '../../api'
import { useAuth } from '../../context/AuthContext'
import toast from 'react-hot-toast'
import { PlusIcon, PencilIcon, TrashIcon, MagnifyingGlassIcon, PhotoIcon } from '@heroicons/react/24/outline'
import ProductForm from './ProductForm'

const Products = () => {
  const { user } = useAuth()
  const [products, setProducts] = useState([])
  const [loading, setLoading] = useState(false)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)
  const [showModal, setShowModal] = useState(false)
  const [editingProduct, setEditingProduct] = useState(null)
  const [imageErrors, setImageErrors] = useState({})
  const limit = 10

  useEffect(() => {
    fetchProducts()
  }, [search, page])

  const fetchProducts = async () => {
    setLoading(true)
    try {
      const response = await api.get('/products', {
        params: { search, page, limit }
      })
      
      console.log('📦 ===== FULL RESPONSE =====')
      console.log('📦 Response data:', response.data)
      
      // Cek struktur response
      // Kemungkinan 1: { data: [...], total: X, total_pages: Y }
      // Kemungkinan 2: { data: [...], pagination: { total: X, total_pages: Y } }
      // Kemungkinan 3: { success: true, data: { ... } }
      // Kemungkinan 4: Langsung array [...]
      
      let productsData = []
      let totalData = 0
      let totalPagesData = 1
      
      // Cek struktur response dari backend
      if (response.data.data && Array.isArray(response.data.data)) {
        // Format: { data: [...], total: X, total_pages: Y }
        productsData = response.data.data
        totalData = response.data.total || response.data.pagination?.total || productsData.length
        totalPagesData = response.data.total_pages || response.data.pagination?.totalPages || 1
        
        console.log('📦 Format 1: data is array in response.data')
      } else if (response.data.success && response.data.data) {
        // Format: { success: true, data: { users: [...], pagination: {...} } }
        const innerData = response.data.data
        if (innerData.products && Array.isArray(innerData.products)) {
          productsData = innerData.products
          totalData = innerData.pagination?.total || productsData.length
          totalPagesData = innerData.pagination?.totalPages || 1
        } else if (innerData.data && Array.isArray(innerData.data)) {
          productsData = innerData.data
          totalData = innerData.total || innerData.pagination?.total || productsData.length
          totalPagesData = innerData.total_pages || innerData.pagination?.totalPages || 1
        } else if (Array.isArray(innerData)) {
          productsData = innerData
          totalData = productsData.length
          totalPagesData = 1
        } else {
          productsData = innerData || []
          totalData = productsData.length
          totalPagesData = 1
        }
        console.log('📦 Format 2: success response')
      } else if (Array.isArray(response.data)) {
        // Format: Langsung array
        productsData = response.data
        totalData = productsData.length
        totalPagesData = 1
        console.log('📦 Format 3: direct array')
      } else if (response.data.pagination) {
        // Format: { data: [...], pagination: { total: X, totalPages: Y } }
        productsData = response.data.data || []
        totalData = response.data.pagination.total || productsData.length
        totalPagesData = response.data.pagination.totalPages || 1
        console.log('📦 Format 4: pagination object')
      } else {
        // Fallback: coba ambil apapun yang bisa
        productsData = response.data.data || response.data || []
        totalData = productsData.length
        totalPagesData = 1
        console.log('📦 Format 5: fallback')
      }
      
      console.log(`📦 Total products: ${productsData.length}`)
      console.log(`📦 Total records: ${totalData}`)
      console.log(`📦 Total pages: ${totalPagesData}`)
      
      // Debug detail untuk setiap product
      productsData.forEach((p, index) => {
        console.log(`📦 Product ${index + 1}: ${p.product_name}`)
        console.log(`   ID: ${p.id}`)
        console.log(`   Image exists: ${!!p.image}`)
        if (p.image) {
          console.log(`   Image length: ${p.image.length}`)
          console.log(`   Image prefix: ${p.image.substring(0, 50)}...`)
          console.log(`   Starts with data:image: ${p.image.startsWith('data:image')}`)
        } else {
          console.log(`   Image: NULL or empty`)
        }
      })
      
      setProducts(productsData)
      setTotal(totalData)
      setTotalPages(totalPagesData)
      
    } catch (error) {
      console.error('❌ Error fetching products:', error)
      toast.error('Failed to fetch products')
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this product?')) return
    
    try {
      await api.delete(`/products/${id}`)
      toast.success('Product deleted successfully')
      fetchProducts()
    } catch (error) {
      toast.error('Failed to delete product')
    }
  }

  const handleExport = async () => {
    try {
      const response = await api.get('/products/export', {
        responseType: 'blob'
      })
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', 'products.xlsx')
      document.body.appendChild(link)
      link.click()
      link.remove()
      toast.success('Products exported successfully')
    } catch (error) {
      toast.error('Failed to export products')
    }
  }

  // Fungsi untuk mendapatkan icon/emoji berdasarkan nama produk
  const getProductIcon = (name) => {
    const icons = {
      'Laptop': '💻',
      'Mouse': '🖱️',
      'Keyboard': '⌨️',
      'Monitor': '🖥️',
      'Printer': '🖨️',
      'Paper': '📄',
      'Pen': '🖊️',
      'USB': '💾',
      'Hard Drive': '💽',
      'Webcam': '📷'
    }
    
    for (const [key, icon] of Object.entries(icons)) {
      if (name?.includes(key)) {
        return icon
      }
    }
    return '📦'
  }

  // Fungsi untuk mendapatkan warna background berdasarkan nama produk
  const getProductColor = (name) => {
    const colorMap = {
      'Laptop': 'bg-blue-500',
      'Mouse': 'bg-green-500',
      'Keyboard': 'bg-red-500',
      'Monitor': 'bg-yellow-500',
      'Printer': 'bg-purple-500',
      'Paper': 'bg-orange-500',
      'Pen': 'bg-pink-500',
      'USB': 'bg-cyan-500',
      'Hard Drive': 'bg-gray-500',
      'Webcam': 'bg-indigo-500'
    }
    
    for (const [key, color] of Object.entries(colorMap)) {
      if (name?.includes(key)) {
        return color
      }
    }
    return 'bg-gray-400'
  }

  // Komponen untuk render gambar dengan fallback icon
  const ProductImage = ({ product }) => {
    const [imgError, setImgError] = useState(false)
    
    // Cek apakah image valid
    const hasImage = product && 
                    product.image && 
                    typeof product.image === 'string' && 
                    product.image.length > 100 &&
                    product.image.startsWith('data:image')

    // Jika gambar error atau tidak ada, tampilkan placeholder dengan icon
    if (imgError || !hasImage) {
      const icon = getProductIcon(product?.product_name || '')
      const colorClass = getProductColor(product?.product_name || '')
      
      return (
        <div className={`w-12 h-12 rounded-lg ${colorClass} flex items-center justify-center text-white text-2xl border border-gray-200 shadow-sm`}>
          {icon}
        </div>
      )
    }

    // Tampilkan gambar
    return (
      <img 
        src={product.image} 
        alt={product.product_name}
        className="w-12 h-12 rounded-lg object-cover border border-gray-200 shadow-sm hover:shadow-md transition-shadow"
        loading="lazy"
        onError={() => {
          console.log(`❌ Image error for: ${product.product_name}`)
          setImgError(true)
        }}
        onLoad={() => {
          console.log(`✅ Image loaded for: ${product.product_name}`)
        }}
      />
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <h2 className="text-2xl font-bold text-gray-900">Products</h2>
        <div className="flex items-center gap-3">
          <button
            onClick={handleExport}
            className="btn-secondary flex items-center gap-2"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            Export Excel
          </button>
          {(user?.role === 'superadmin' || user?.role === 'head') && (
            <button
              onClick={() => {
                setEditingProduct(null)
                setShowModal(true)
              }}
              className="btn-primary flex items-center gap-2"
            >
              <PlusIcon className="h-5 w-5" />
              Add Product
            </button>
          )}
        </div>
      </div>

      {/* Search */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search products..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input-field pl-10"
          />
        </div>
      </div>

      {/* Table */}
      <div className="card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr>
                <th className="table-header">Image</th>
                <th className="table-header">Code</th>
                <th className="table-header">Name</th>
                <th className="table-header">Category</th>
                <th className="table-header">Unit</th>
                <th className="table-header">Min Stock</th>
                <th className="table-header">Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan="7" className="text-center py-8">
                    <div className="flex justify-center">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
                    </div>
                  </td>
                </tr>
              ) : products.length === 0 ? (
                <tr>
                  <td colSpan="7" className="text-center py-8 text-gray-500">
                    No products found
                  </td>
                </tr>
              ) : (
                products.map((product) => (
                  <tr key={product.id} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
                    <td className="table-cell">
                      <ProductImage product={product} />
                    </td>
                    <td className="table-cell font-medium">{product.product_code}</td>
                    <td className="table-cell">{product.product_name}</td>
                    <td className="table-cell">{product.category || '-'}</td>
                    <td className="table-cell">{product.unit}</td>
                    <td className="table-cell">{product.min_stock}</td>
                    <td className="table-cell">
                      <div className="flex items-center gap-2">
                        {(user?.role === 'superadmin' || user?.role === 'head') && (
                          <button
                            onClick={() => {
                              setEditingProduct(product)
                              setShowModal(true)
                            }}
                            className="p-1 text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                            title="Edit product"
                          >
                            <PencilIcon className="h-5 w-5" />
                          </button>
                        )}
                        {user?.role === 'superadmin' && (
                          <button
                            onClick={() => handleDelete(product.id)}
                            className="p-1 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                            title="Delete product"
                          >
                            <TrashIcon className="h-5 w-5" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between px-6 py-3 border-t border-gray-200">
            <div className="text-sm text-gray-600">
              Showing {((page - 1) * limit) + 1} to {Math.min(page * limit, total)} of {total} results
            </div>
            
            <div className="flex items-center gap-2">
              {/* Tombol First */}
              <button
                onClick={() => setPage(1)}
                disabled={page === 1}
                className="px-3 py-1 rounded-lg border border-gray-300 disabled:opacity-50 hover:bg-gray-50 transition-colors"
              >
                <span className="flex items-center gap-1">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 19l-7-7 7-7" />
                  </svg>
                  First
                </span>
              </button>

              {/* Tombol Previous */}
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 rounded-lg border border-gray-300 disabled:opacity-50 hover:bg-gray-50 transition-colors"
              >
                <span className="flex items-center gap-1">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                  </svg>
                  Prev
                </span>
              </button>

              {/* Nomor Halaman dengan Ellipsis */}
              <div className="flex gap-1">
                {(() => {
                  const pages = [];
                  const maxVisible = 7;
                  const halfVisible = Math.floor(maxVisible / 2);
                  
                  let startPage = Math.max(1, page - halfVisible);
                  let endPage = Math.min(totalPages, page + halfVisible);
                  
                  if (page <= halfVisible + 1) {
                    endPage = Math.min(totalPages, maxVisible);
                  }
                  
                  if (page > totalPages - halfVisible) {
                    startPage = Math.max(1, totalPages - maxVisible + 1);
                  }

                  if (startPage > 1) {
                    pages.push(1);
                    if (startPage > 2) {
                      pages.push('ellipsis');
                    }
                  }

                  for (let i = startPage; i <= endPage; i++) {
                    pages.push(i);
                  }

                  if (endPage < totalPages) {
                    if (endPage < totalPages - 1) {
                      pages.push('ellipsis');
                    }
                    pages.push(totalPages);
                  }

                  return pages.map((p, index) => {
                    if (p === 'ellipsis') {
                      return (
                        <span key={`ellipsis-${index}`} className="px-3 py-1 text-gray-500 select-none">
                          …
                        </span>
                      );
                    }
                    
                    return (
                      <button
                        key={p}
                        onClick={() => setPage(p)}
                        className={`min-w-[36px] h-9 px-3 rounded-lg border transition-colors ${
                          page === p
                            ? 'bg-blue-600 text-white border-blue-600 hover:bg-blue-700'
                            : 'border-gray-300 hover:bg-gray-50'
                        }`}
                      >
                        {p}
                      </button>
                    );
                  });
                })()}
              </div>

              {/* Tombol Next */}
              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-3 py-1 rounded-lg border border-gray-300 disabled:opacity-50 hover:bg-gray-50 transition-colors"
              >
                <span className="flex items-center gap-1">
                  Next
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </span>
              </button>

              {/* Tombol Last */}
              <button
                onClick={() => setPage(totalPages)}
                disabled={page === totalPages}
                className="px-3 py-1 rounded-lg border border-gray-300 disabled:opacity-50 hover:bg-gray-50 transition-colors"
              >
                <span className="flex items-center gap-1">
                  Last
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 5l7 7-7 7" />
                  </svg>
                </span>
              </button>

              {/* Go to Page dengan Dropdown */}
              <div className="flex items-center gap-2 ml-2 border-l border-gray-200 pl-3">
                <span className="text-sm text-gray-600">Go to</span>
                <select
                  value={page}
                  onChange={(e) => setPage(Number(e.target.value))}
                  className="px-2 py-1 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                >
                  {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                    <option key={p} value={p}>
                      {p}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <ProductForm
          product={editingProduct}
          onClose={() => {
            setShowModal(false)
            setEditingProduct(null)
          }}
          onSuccess={fetchProducts}
        />
      )}
    </div>
  )
}

export default Products