import React, { useState, useEffect } from 'react'
import api from '../../api'
import toast from 'react-hot-toast'
import { MagnifyingGlassIcon, ArrowDownTrayIcon, FunnelIcon } from '@heroicons/react/24/outline'

const Stock = () => {
  const [stock, setStock] = useState([])
  const [loading, setLoading] = useState(false)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)
  const limit = 10

  useEffect(() => {
    fetchStock()
  }, [search, statusFilter, page])

  const fetchStock = async () => {
    setLoading(true)
    try {
      const response = await api.get('/stock', {
        params: { search, status: statusFilter, page, limit }
      })
      
      console.log('📦 ===== STOCK RESPONSE =====')
      console.log('📦 Response data:', response.data)
      
      let stockData = []
      let totalData = 0
      let totalPagesData = 1
      
      // Format 1: { data: [...], total: X, total_pages: Y }
      if (response.data.data && Array.isArray(response.data.data)) {
        stockData = response.data.data
        totalData = response.data.total || response.data.pagination?.total || stockData.length
        totalPagesData = response.data.total_pages || response.data.pagination?.totalPages || 1
        console.log('📦 Format 1: data in response.data')
      } 
      // Format 2: { success: true, data: { data: [...], pagination: {...} } }
      else if (response.data.success && response.data.data) {
        const innerData = response.data.data
        if (innerData.data && Array.isArray(innerData.data)) {
          stockData = innerData.data
          totalData = innerData.pagination?.total || innerData.total || stockData.length
          totalPagesData = innerData.pagination?.totalPages || innerData.total_pages || 1
        } else if (Array.isArray(innerData)) {
          stockData = innerData
          totalData = stockData.length
          totalPagesData = 1
        } else {
          stockData = innerData || []
          totalData = stockData.length
          totalPagesData = 1
        }
        console.log('📦 Format 2: success response')
      } 
      // Format 3: Langsung array
      else if (Array.isArray(response.data)) {
        stockData = response.data
        totalData = stockData.length
        totalPagesData = 1
        console.log('📦 Format 3: direct array')
      } 
      // Format 4: { data: [...], pagination: {...} }
      else if (response.data.pagination) {
        stockData = response.data.data || []
        totalData = response.data.pagination.total || stockData.length
        totalPagesData = response.data.pagination.totalPages || 1
        console.log('📦 Format 4: pagination object')
      } 
      // Fallback
      else {
        stockData = response.data.data || response.data || []
        totalData = stockData.length
        totalPagesData = 1
        console.log('📦 Format 5: fallback')
      }
      
      console.log(`📦 Total stock: ${stockData.length}`)
      console.log(`📦 Total records: ${totalData}`)
      console.log(`📦 Total pages: ${totalPagesData}`)
      
      if (stockData.length > 0) {
        console.log('📦 First item structure:', Object.keys(stockData[0]))
        console.log('📦 First item:', stockData[0])
      }
      
      setStock(stockData)
      setTotal(totalData)
      setTotalPages(totalPagesData)
      
    } catch (error) {
      console.error('❌ Error fetching stock:', error)
      toast.error('Failed to fetch stock')
    } finally {
      setLoading(false)
    }
  }

  const handleExport = async () => {
    try {
      const response = await api.get('/stock/export', {
        responseType: 'blob'
      })
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', 'stock.xlsx')
      document.body.appendChild(link)
      link.click()
      link.remove()
      toast.success('Stock exported successfully')
    } catch (error) {
      toast.error('Failed to export stock')
    }
  }

  const getStatusBadge = (quantity, minStock) => {
    if (quantity === 0) return 'badge-danger'
    if (quantity <= minStock) return 'badge-warning'
    return 'badge-success'
  }

  const getStatusText = (quantity, minStock) => {
    if (quantity === 0) return 'Out of Stock'
    if (quantity <= minStock) return 'Low Stock'
    return 'Normal'
  }

  // Helper untuk mendapatkan data dari item
  const getItemData = (item) => {
    if (item.product) {
      return {
        id: item.id,
        productCode: item.product.product_code || '-',
        productName: item.product.product_name || '-',
        category: item.product.category || '-',
        unit: item.product.unit || '-',
        quantity: item.quantity || 0,
        minStock: item.product.min_stock || 0,
        status: getStatusText(item.quantity || 0, item.product.min_stock || 0)
      }
    }
    return {
      id: item.id,
      productCode: item.product_code || item.productCode || '-',
      productName: item.product_name || item.productName || '-',
      category: item.category || '-',
      unit: item.unit || '-',
      quantity: item.quantity || item.current_stock || 0,
      minStock: item.min_stock || item.minStock || 0,
      status: getStatusText(item.quantity || 0, item.min_stock || 0)
    }
  }

  // Filter stock berdasarkan status
  const filteredStock = stock.filter(item => {
    const data = getItemData(item)
    if (!statusFilter) return true
    return data.status === statusFilter
  })

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <h2 className="text-2xl font-bold text-gray-900">Stock Management</h2>
        <button
          onClick={handleExport}
          className="btn-secondary flex items-center gap-2"
        >
          <ArrowDownTrayIcon className="h-5 w-5" />
          Export Excel
        </button>
      </div>

      {/* Search & Filter */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search stock by product name or code..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input-field pl-10"
          />
        </div>
        <div className="flex items-center gap-2">
          <FunnelIcon className="h-5 w-5 text-gray-400" />
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value)
              setPage(1)
            }}
            className="input-field w-48"
          >
            <option value="">All Status</option>
            <option value="Normal">✅ Normal</option>
            <option value="Low Stock">⚠️ Low Stock</option>
            <option value="Out of Stock">❌ Out of Stock</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr>
                <th className="table-header">Product Code</th>
                <th className="table-header">Product Name</th>
                <th className="table-header">Category</th>
                <th className="table-header">Unit</th>
                <th className="table-header">Current Stock</th>
                <th className="table-header">Min Stock</th>
                <th className="table-header">Status</th>
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
              ) : filteredStock.length === 0 ? (
                <tr>
                  <td colSpan="7" className="text-center py-8 text-gray-500">
                    No stock data found
                  </td>
                </tr>
              ) : (
                filteredStock.map((item) => {
                  const data = getItemData(item)
                  return (
                    <tr key={data.id} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
                      <td className="table-cell font-medium">{data.productCode}</td>
                      <td className="table-cell">{data.productName}</td>
                      <td className="table-cell">{data.category}</td>
                      <td className="table-cell">{data.unit}</td>
                      <td className="table-cell font-semibold">{data.quantity}</td>
                      <td className="table-cell">{data.minStock}</td>
                      <td className="table-cell">
                        <span className={`badge ${getStatusBadge(data.quantity, data.minStock)}`}>
                          {data.status}
                        </span>
                      </td>
                    </tr>
                  )
                })
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
    </div>
  )
}

export default Stock