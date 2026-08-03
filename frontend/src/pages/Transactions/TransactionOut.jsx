import React, { useState, useEffect, useRef } from 'react'
import api from '../../api'
import { useAuth } from '../../context/AuthContext'
import toast from 'react-hot-toast'
import { 
  MagnifyingGlassIcon, 
  PlusIcon, 
  CheckIcon, 
  XMarkIcon, 
  ArrowsUpDownIcon, 
  ArrowDownTrayIcon,
  ChevronLeftIcon,
  ChevronRightIcon
} from '@heroicons/react/24/outline'
import { format } from 'date-fns'

const TransactionOut = () => {
  const { user } = useAuth()
  const [transactions, setTransactions] = useState([])
  const [products, setProducts] = useState([])
  const [loading, setLoading] = useState(false)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [sortOrder, setSortOrder] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState({
    product_id: '',
    quantity: '',
    note: ''
  })
  const [submitting, setSubmitting] = useState(false)
  const limit = 10

  // === FORM STATE ===
  const [productSearch, setProductSearch] = useState('')
  const [productPage, setProductPage] = useState(1)
  const [filteredProducts, setFilteredProducts] = useState([])
  const [totalProducts, setTotalProducts] = useState(0)
  const [totalProductPages, setTotalProductPages] = useState(1)
  const [selectedProduct, setSelectedProduct] = useState(null)
  const [showDropdown, setShowDropdown] = useState(false)
  const productLimit = 10
  const dropdownRef = useRef(null)
  const searchInputRef = useRef(null)

  const canApprove = user?.role === 'superadmin' || user?.role === 'head'

  useEffect(() => {
    if (sortOrder) {
      fetchAllTransactions()
    } else {
      fetchTransactions()
    }
    fetchProducts()
  }, [search, statusFilter, sortOrder, page])

  // === FETCH PRODUCTS FOR DROPDOWN ===
  useEffect(() => {
    if (showForm) {
      fetchFilteredProducts()
    }
  }, [productSearch, productPage, showForm])

  // === CLOSE DROPDOWN ON CLICK OUTSIDE ===
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setShowDropdown(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const fetchFilteredProducts = async () => {
    try {
      const response = await api.get('/products', {
        params: { 
          search: productSearch, 
          page: productPage, 
          limit: productLimit 
        }
      })
      
      let productsData = []
      let totalData = 0
      
      if (response.data.data && Array.isArray(response.data.data)) {
        productsData = response.data.data
        totalData = response.data.total || response.data.pagination?.total || productsData.length
      } else if (response.data.success && response.data.data) {
        const innerData = response.data.data
        if (innerData.data && Array.isArray(innerData.data)) {
          productsData = innerData.data
          totalData = innerData.pagination?.total || innerData.total || productsData.length
        } else if (Array.isArray(innerData)) {
          productsData = innerData
          totalData = productsData.length
        } else {
          productsData = innerData || []
          totalData = productsData.length
        }
      } else if (Array.isArray(response.data)) {
        productsData = response.data
        totalData = productsData.length
      } else {
        productsData = response.data.data || response.data || []
        totalData = productsData.length
      }
      
      setFilteredProducts(productsData)
      setTotalProducts(totalData)
      setTotalProductPages(Math.ceil(totalData / productLimit))
    } catch (error) {
      console.error('Failed to fetch products:', error)
    }
  }

  const fetchTransactions = async () => {
    setLoading(true)
    try {
      const response = await api.get('/transactions/out', {
        params: { search, status: statusFilter, page, limit }
      })
      
      let transactionsData = []
      let totalData = 0
      let totalPagesData = 1
      
      if (response.data.data && Array.isArray(response.data.data)) {
        transactionsData = response.data.data
        totalData = response.data.total || response.data.pagination?.total || transactionsData.length
        totalPagesData = response.data.total_pages || response.data.pagination?.totalPages || 1
      } else if (response.data.success && response.data.data) {
        const innerData = response.data.data
        if (innerData.data && Array.isArray(innerData.data)) {
          transactionsData = innerData.data
          totalData = innerData.pagination?.total || innerData.total || transactionsData.length
          totalPagesData = innerData.pagination?.totalPages || innerData.total_pages || 1
        } else if (Array.isArray(innerData)) {
          transactionsData = innerData
          totalData = transactionsData.length
          totalPagesData = 1
        } else {
          transactionsData = innerData || []
          totalData = transactionsData.length
          totalPagesData = 1
        }
      } else if (Array.isArray(response.data)) {
        transactionsData = response.data
        totalData = transactionsData.length
        totalPagesData = 1
      } else if (response.data.pagination) {
        transactionsData = response.data.data || []
        totalData = response.data.pagination.total || transactionsData.length
        totalPagesData = response.data.pagination.totalPages || 1
      } else {
        transactionsData = response.data.data || response.data || []
        totalData = transactionsData.length
        totalPagesData = 1
      }
      
      setTransactions(transactionsData)
      setTotal(totalData)
      setTotalPages(totalPagesData)
      
    } catch (error) {
      console.error('❌ Error fetching transactions:', error)
      toast.error('Failed to fetch transactions')
    } finally {
      setLoading(false)
    }
  }

  const fetchAllTransactions = async () => {
    setLoading(true)
    try {
      const response = await api.get('/transactions/out', {
        params: { search, status: statusFilter, limit: 9999 }
      })
      
      let transactionsData = []
      let totalData = 0
      
      if (response.data.data && Array.isArray(response.data.data)) {
        transactionsData = response.data.data
        totalData = response.data.total || response.data.pagination?.total || transactionsData.length
      } else if (response.data.success && response.data.data) {
        const innerData = response.data.data
        if (innerData.data && Array.isArray(innerData.data)) {
          transactionsData = innerData.data
          totalData = innerData.pagination?.total || innerData.total || transactionsData.length
        } else if (Array.isArray(innerData)) {
          transactionsData = innerData
          totalData = transactionsData.length
        } else {
          transactionsData = innerData || []
          totalData = transactionsData.length
        }
      } else if (Array.isArray(response.data)) {
        transactionsData = response.data
        totalData = transactionsData.length
      } else {
        transactionsData = response.data.data || response.data || []
        totalData = transactionsData.length
      }
      
      let sortedData = [...transactionsData]
      if (sortOrder === 'desc') {
        sortedData.sort((a, b) => (b.quantity || 0) - (a.quantity || 0))
      } else if (sortOrder === 'asc') {
        sortedData.sort((a, b) => (a.quantity || 0) - (b.quantity || 0))
      }
      
      const totalRecords = sortedData.length
      const totalPagesData = Math.ceil(totalRecords / limit)
      const startIndex = (page - 1) * limit
      const endIndex = startIndex + limit
      const paginatedData = sortedData.slice(startIndex, endIndex)
      
      setTransactions(paginatedData)
      setTotal(totalRecords)
      setTotalPages(totalPagesData)
      
    } catch (error) {
      console.error('❌ Error fetching transactions for sorting:', error)
      toast.error('Failed to fetch transactions')
      fetchTransactions()
    } finally {
      setLoading(false)
    }
  }

  const fetchProducts = async () => {
    try {
      const response = await api.get('/products', { params: { limit: 100 } })
      let productsData = []
      if (response.data.data && Array.isArray(response.data.data)) {
        productsData = response.data.data
      } else if (Array.isArray(response.data)) {
        productsData = response.data
      } else {
        productsData = response.data.data || response.data || []
      }
      setProducts(productsData)
    } catch (error) {
      console.error('Failed to fetch products')
    }
  }

  const handleExport = async () => {
    try {
      const loadingToast = toast.loading('Exporting transactions...')
      
      const response = await api.get('/transactions/out/export', {
        params: { 
          search: search || undefined,
          status: statusFilter || undefined
        },
        responseType: 'blob'
      })
      
      toast.dismiss(loadingToast)
      
      if (!response.data || response.data.size === 0) {
        toast.error('No data to export')
        return
      }
      
      const blob = new Blob([response.data], { 
        type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' 
      })
      const url = window.URL.createObjectURL(blob)
      
      const link = document.createElement('a')
      link.href = url
      const filename = `transaction_out_${format(new Date(), 'yyyy-MM-dd_HH-mm')}.xlsx`
      link.setAttribute('download', filename)
      document.body.appendChild(link)
      link.click()
      link.remove()
      
      setTimeout(() => window.URL.revokeObjectURL(url), 100)
      
      toast.success('Transactions exported successfully!')
    } catch (error) {
      toast.dismiss()
      console.error('❌ Export error:', error)
      toast.error(error.response?.data?.message || 'Failed to export transactions')
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSubmitting(true)

    try {
      await api.post('/transactions/out', {
        product_id: parseInt(formData.product_id),
        quantity: parseInt(formData.quantity),
        note: formData.note
      })
      toast.success('Transaction out created successfully')
      setShowForm(false)
      setFormData({ product_id: '', quantity: '', note: '' })
      setSelectedProduct(null)
      setProductSearch('')
      if (sortOrder) {
        fetchAllTransactions()
      } else {
        fetchTransactions()
      }
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failed to create transaction')
    } finally {
      setSubmitting(false)
    }
  }

  const handleApprove = async (id, status) => {
    if (!window.confirm(`Are you sure you want to ${status} this transaction?`)) return

    try {
      await api.put(`/transactions/out/${id}/approve`, { status })
      toast.success(`Transaction ${status} successfully`)
      if (sortOrder) {
        fetchAllTransactions()
      } else {
        fetchTransactions()
      }
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failed to approve transaction')
    }
  }

  const toggleSort = () => {
    let newSortOrder = ''
    if (sortOrder === 'desc') {
      newSortOrder = 'asc'
    } else if (sortOrder === 'asc') {
      newSortOrder = ''
    } else {
      newSortOrder = 'desc'
    }
    setSortOrder(newSortOrder)
    setPage(1)
  }

  const getSortLabel = () => {
    if (sortOrder === 'desc') return 'Quantity: Highest ↓'
    if (sortOrder === 'asc') return 'Quantity: Lowest ↑'
    return 'Sort by Quantity'
  }

  const getStatusBadge = (status) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800'
      case 'approved':
        return 'bg-green-100 text-green-800'
      case 'rejected':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusText = (status) => {
    switch (status) {
      case 'pending':
        return '⏳ Pending'
      case 'approved':
        return '✅ Approved'
      case 'rejected':
        return '❌ Rejected'
      default:
        return status
    }
  }

  const getTransactionData = (item) => {
    if (item.product) {
      return {
        id: item.id,
        productName: item.product.product_name || '-',
        productCode: item.product.product_code || '-',
        quantity: item.quantity || 0,
        status: item.status || 'pending',
        note: item.note || '-',
        createdBy: item.created_by_name || item.created_by || '-',
        createdAt: item.created_at
      }
    }
    return {
      id: item.id,
      productName: item.product_name || item.product?.product_name || '-',
      productCode: item.product_code || item.product?.product_code || '-',
      quantity: item.quantity || 0,
      status: item.status || 'pending',
      note: item.note || '-',
      createdBy: item.created_by_name || item.created_by || '-',
      createdAt: item.created_at
    }
  }

  // === SELECT PRODUCT ===
  const handleSelectProduct = (product) => {
    setSelectedProduct(product)
    setFormData({ ...formData, product_id: product.id })
    setProductSearch(product.product_name)
    setShowDropdown(false)
  }

  // === CLEAR SELECTED PRODUCT ===
  const clearSelectedProduct = () => {
    setSelectedProduct(null)
    setFormData({ ...formData, product_id: '' })
    setProductSearch('')
    setShowDropdown(true)
    setTimeout(() => searchInputRef.current?.focus(), 100)
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <h2 className="text-2xl font-bold text-gray-900">Transaction Out</h2>
        <div className="flex items-center gap-3">
          <button
            onClick={handleExport}
            className="btn-secondary flex items-center gap-2"
          >
            <ArrowDownTrayIcon className="h-5 w-5" />
            Export Excel
          </button>
          <button
            onClick={() => {
              setShowForm(true)
              setProductSearch('')
              setSelectedProduct(null)
              setFormData({ product_id: '', quantity: '', note: '' })
              setProductPage(1)
              setTimeout(() => searchInputRef.current?.focus(), 300)
            }}
            className="btn-primary flex items-center gap-2"
          >
            <PlusIcon className="h-5 w-5" />
            Request Out
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search transactions..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input-field pl-10"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="input-field sm:w-48"
        >
          <option value="">All Status</option>
          <option value="pending">⏳ Pending</option>
          <option value="approved">✅ Approved</option>
          <option value="rejected">❌ Rejected</option>
        </select>
        <button
          onClick={toggleSort}
          className={`px-4 py-2 rounded-lg border transition-colors flex items-center gap-2 whitespace-nowrap ${
            sortOrder ? 'bg-blue-50 border-blue-500 text-blue-700' : 'border-gray-300 hover:bg-gray-50'
          }`}
        >
          <ArrowsUpDownIcon className="h-5 w-5" />
          {getSortLabel()}
          {sortOrder && (
            <span className="text-xs text-gray-500 ml-1">
              (Global)
            </span>
          )}
        </button>
      </div>

      {/* Table */}
      <div className="card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr>
                <th className="table-header">Product</th>
                <th className="table-header">
                  <button
                    onClick={toggleSort}
                    className="flex items-center gap-1 hover:text-blue-600 transition-colors"
                  >
                    Quantity
                    {sortOrder === 'desc' && <span className="text-xs">↓</span>}
                    {sortOrder === 'asc' && <span className="text-xs">↑</span>}
                  </button>
                </th>
                <th className="table-header">Status</th>
                <th className="table-header">Note</th>
                <th className="table-header">Requested By</th>
                <th className="table-header">Date</th>
                {canApprove && <th className="table-header">Actions</th>}
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
              ) : transactions.length === 0 ? (
                <tr>
                  <td colSpan="7" className="text-center py-8 text-gray-500">
                    No transactions found
                  </td>
                </tr>
              ) : (
                transactions.map((item) => {
                  const data = getTransactionData(item)
                  const requesterName = data.createdBy || `User #${item.created_by}`
                  
                  return (
                    <tr key={data.id} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
                      <td className="table-cell">
                        <div>
                          <div className="font-medium">{data.productName}</div>
                          <div className="text-xs text-gray-500">{data.productCode}</div>
                        </div>
                      </td>
                      <td className="table-cell">
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-red-600">
                            {data.quantity}
                          </span>
                          <span className="text-xs font-medium text-red-500 bg-red-50 px-2 py-0.5 rounded-full">
                            OUT
                          </span>
                        </div>
                      </td>
                      <td className="table-cell">
                        <span className={`px-2 py-1 text-xs font-semibold rounded-full ${getStatusBadge(data.status)}`}>
                          {getStatusText(data.status)}
                        </span>
                      </td>
                      <td className="table-cell">{data.note}</td>
                      <td className="table-cell">
                        <div className="flex items-center gap-2">
                          <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 text-sm font-semibold">
                            {requesterName.charAt(0).toUpperCase()}
                          </div>
                          <span className="font-medium">{requesterName}</span>
                        </div>
                      </td>
                      <td className="table-cell text-sm">
                        {data.createdAt ? format(new Date(data.createdAt), 'dd/MM/yyyy HH:mm') : '-'}
                      </td>
                      {canApprove && data.status === 'pending' && (
                        <td className="table-cell">
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => handleApprove(data.id, 'approved')}
                              className="p-1.5 bg-green-100 text-green-600 hover:bg-green-200 rounded-lg transition-colors"
                              title="Approve"
                            >
                              <CheckIcon className="h-5 w-5" />
                            </button>
                            <button
                              onClick={() => handleApprove(data.id, 'rejected')}
                              className="p-1.5 bg-red-100 text-red-600 hover:bg-red-200 rounded-lg transition-colors"
                              title="Reject"
                            >
                              <XMarkIcon className="h-5 w-5" />
                            </button>
                          </div>
                        </td>
                      )}
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
              {sortOrder && <span className="text-xs text-blue-600 ml-2">(Sorted globally)</span>}
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

      {/* ==================== SUPER MODAL ==================== */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full max-h-[90vh] overflow-hidden">
            <div className="p-6 border-b border-gray-200 flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold">Request Stock Out</h3>
                <p className="text-sm text-gray-500 mt-0.5">
                  {user?.role === 'head' || user?.role === 'superadmin' 
                    ? '✅ Request will be auto-approved' 
                    : '⏳ Request will be sent for approval'}
                </p>
              </div>
              <button
                onClick={() => {
                  setShowForm(false)
                  setSelectedProduct(null)
                  setProductSearch('')
                  setFormData({ product_id: '', quantity: '', note: '' })
                }}
                className="p-1 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <XMarkIcon className="h-6 w-6" />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="p-6 space-y-4 overflow-y-auto max-h-[calc(90vh-120px)]">
              {/* Product Selection - SUPER DROPDOWN */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Product <span className="text-red-500">*</span>
                </label>
                <div ref={dropdownRef} className="relative">
                  <div className="flex gap-2">
                    <div className="relative flex-1">
                      <input
                        ref={searchInputRef}
                        type="text"
                        value={productSearch}
                        onChange={(e) => {
                          setProductSearch(e.target.value)
                          setShowDropdown(true)
                          setProductPage(1)
                          if (e.target.value === '') {
                            setSelectedProduct(null)
                            setFormData({ ...formData, product_id: '' })
                          }
                        }}
                        onFocus={() => setShowDropdown(true)}
                        placeholder="Search product by name or code..."
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 pr-8"
                      />
                      {productSearch && (
                        <button
                          type="button"
                          onClick={clearSelectedProduct}
                          className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                        >
                          <XMarkIcon className="h-5 w-5" />
                        </button>
                      )}
                    </div>
                  </div>

                  {/* Dropdown List */}
                  {showDropdown && (
                    <div className="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg max-h-64 overflow-hidden">
                      <div className="p-2 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                        <span className="text-xs text-gray-500">
                          {totalProducts} products found
                        </span>
                      </div>
                      <div className="overflow-y-auto max-h-48">
                        {filteredProducts.length === 0 ? (
                          <div className="p-4 text-center text-gray-500 text-sm">
                            No products found
                          </div>
                        ) : (
                          filteredProducts.map((product) => (
                            <button
                              key={product.id}
                              type="button"
                              onClick={() => handleSelectProduct(product)}
                              className={`w-full px-3 py-2 text-left hover:bg-blue-50 transition-colors flex items-center justify-between ${
                                selectedProduct?.id === product.id ? 'bg-blue-50' : ''
                              }`}
                            >
                              <div className="flex-1 min-w-0">
                                <div className="text-sm font-medium text-gray-900 truncate">
                                  {product.product_name}
                                </div>
                                <div className="text-xs text-gray-500 truncate">
                                  {product.product_code} • {product.category || 'No category'}
                                </div>
                              </div>
                              {selectedProduct?.id === product.id && (
                                <span className="text-blue-600 text-sm font-medium ml-2">✓</span>
                              )}
                            </button>
                          ))
                        )}
                      </div>
                      
                      {/* Pagination for products */}
                      {totalProductPages > 1 && (
                        <div className="flex items-center justify-between p-2 border-t border-gray-100 bg-gray-50">
                          <button
                            type="button"
                            onClick={() => setProductPage(p => Math.max(1, p - 1))}
                            disabled={productPage === 1}
                            className="p-1 rounded hover:bg-gray-200 disabled:opacity-50"
                          >
                            <ChevronLeftIcon className="h-4 w-4" />
                          </button>
                          <span className="text-xs text-gray-500">
                            Page {productPage} of {totalProductPages}
                          </span>
                          <button
                            type="button"
                            onClick={() => setProductPage(p => Math.min(totalProductPages, p + 1))}
                            disabled={productPage === totalProductPages}
                            className="p-1 rounded hover:bg-gray-200 disabled:opacity-50"
                          >
                            <ChevronRightIcon className="h-4 w-4" />
                          </button>
                        </div>
                      )}
                    </div>
                  )}
                </div>
                
                {/* Selected product display */}
                {selectedProduct && (
                  <div className="mt-2 p-2 bg-blue-50 border border-blue-200 rounded-lg flex items-center justify-between">
                    <div>
                      <span className="text-sm font-medium text-blue-700">{selectedProduct.product_name}</span>
                      <span className="text-xs text-blue-500 ml-2">({selectedProduct.product_code})</span>
                    </div>
                    <span className="text-xs text-blue-600">✓ Selected</span>
                  </div>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Quantity <span className="text-red-500">*</span>
                </label>
                <input
                  type="number"
                  value={formData.quantity}
                  onChange={(e) => setFormData({ ...formData, quantity: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  min="1"
                  required
                  autoFocus={selectedProduct !== null}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Note (Reason) <span className="text-red-500">*</span>
                </label>
                <textarea
                  value={formData.note}
                  onChange={(e) => setFormData({ ...formData, note: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  rows="3"
                  placeholder="Reason for stock out..."
                  required
                />
              </div>

              {/* Stock Info */}
              {selectedProduct && (
                <div className="p-3 bg-gray-50 rounded-lg border border-gray-200">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-600">Selected Product:</span>
                    <span className="font-medium text-gray-900">{selectedProduct.product_name}</span>
                  </div>
                  <div className="flex items-center justify-between text-sm mt-1">
                    <span className="text-gray-600">Current Stock:</span>
                    <span className="font-medium text-blue-600">
                      {selectedProduct.stock?.quantity || 0} units
                    </span>
                  </div>
                </div>
              )}

              <div className="flex gap-3 pt-4">
                <button
                  type="submit"
                  className="btn-primary flex-1"
                  disabled={submitting || !formData.product_id || !formData.quantity || !formData.note}
                >
                  {submitting ? 'Submitting...' : 'Submit Request'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowForm(false)
                    setSelectedProduct(null)
                    setProductSearch('')
                    setFormData({ product_id: '', quantity: '', note: '' })
                  }}
                  className="btn-secondary flex-1"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

export default TransactionOut