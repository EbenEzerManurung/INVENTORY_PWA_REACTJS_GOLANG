import React, { useState, useEffect } from 'react'
import api from '../../api'
import toast from 'react-hot-toast'
import { 
  ArrowDownTrayIcon, 
  ChartBarIcon, 
  MagnifyingGlassIcon,
  ArrowsUpDownIcon,
  FunnelIcon,
  ChevronDoubleLeftIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronDoubleRightIcon
} from '@heroicons/react/24/outline'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
} from 'chart.js'
import { Bar, Pie } from 'react-chartjs-2'

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
)

const Reports = () => {
  const [allReports, setAllReports] = useState([])
  const [reports, setReports] = useState([])
  const [summary, setSummary] = useState(null)
  const [loading, setLoading] = useState(false)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [sortOrder, setSortOrder] = useState('desc')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)
  const [chartData, setChartData] = useState(null)
  const [pieChartData, setPieChartData] = useState(null)
  const limit = 10

  useEffect(() => {
    fetchAllReports()
  }, [])

  useEffect(() => {
    processData()
  }, [allReports, search, statusFilter, sortOrder, page])

  const fetchAllReports = async () => {
    setLoading(true)
    try {
      const response = await api.get('/reports/stock')
      console.log('📦 Reports response:', response.data)
      
      const responseData = response.data.data || response.data
      const reportsData = responseData.data || responseData || []
      const summaryData = responseData.summary || null
      
      const reportsArray = Array.isArray(reportsData) ? reportsData : []
      
      // ✅ Hitung ulang summary di frontend untuk memastikan akurat
      let totalProducts = reportsArray.length
      let totalStock = 0
      let lowStock = 0
      let outOfStock = 0
      let healthy = 0
      
      reportsArray.forEach(item => {
        const quantity = item.quantity || 0
        const minStock = item.min_stock || 0
        
        totalStock += quantity
        
        // Tentukan status berdasarkan quantity dan min_stock
        let status = 'Healthy'
        if (quantity === 0) {
          status = 'Out of Stock'
          outOfStock++
        } else if (quantity <= minStock) {
          status = 'Low Stock'
          lowStock++
        } else {
          healthy++
        }
        
        // Tambahkan status ke item untuk digunakan nanti
        item._status = status
      })
      
      // Update summary dengan data yang akurat
      const accurateSummary = {
        total_products: totalProducts,
        total_stock: totalStock,
        low_stock: lowStock,
        out_of_stock: outOfStock,
        healthy: healthy
      }
      
      setSummary(accurateSummary)
      setAllReports(reportsArray)
      setTotal(reportsArray.length)
      setTotalPages(Math.ceil(reportsArray.length / limit))
      
      // ✅ Prepare chart data
      if (reportsArray.length > 0) {
        // Bar chart - Top 20 by stock
        const sortedForChart = [...reportsArray].sort((a, b) => (b.quantity || 0) - (a.quantity || 0))
        const top20 = sortedForChart.slice(0, 20)
        
        setChartData({
          labels: top20.map(r => r.product_name || 'Unknown'),
          datasets: [
            {
              label: 'Current Stock',
              data: top20.map(r => r.quantity || 0),
              backgroundColor: 'rgba(99, 102, 241, 0.8)',
              borderColor: 'rgba(99, 102, 241, 1)',
              borderWidth: 1,
              borderRadius: 4,
            },
            {
              label: 'Total In',
              data: top20.map(r => r.total_in || 0),
              backgroundColor: 'rgba(52, 211, 153, 0.8)',
              borderColor: 'rgba(52, 211, 153, 1)',
              borderWidth: 1,
              borderRadius: 4,
            },
            {
              label: 'Total Out',
              data: top20.map(r => r.total_out || 0),
              backgroundColor: 'rgba(251, 146, 60, 0.8)',
              borderColor: 'rgba(251, 146, 60, 1)',
              borderWidth: 1,
              borderRadius: 4,
            }
          ]
        })

        // ✅ Pie chart berdasarkan status yang sudah dihitung
        setPieChartData({
          labels: ['Healthy', 'Low Stock', 'Out of Stock'],
          datasets: [
            {
              data: [healthy, lowStock, outOfStock],
              backgroundColor: ['#22c55e', '#eab308', '#ef4444'],
              borderColor: ['#ffffff', '#ffffff', '#ffffff'],
              borderWidth: 2,
            }
          ]
        })
      }
    } catch (error) {
      console.error('❌ Fetch reports error:', error)
      toast.error('Failed to fetch reports')
    } finally {
      setLoading(false)
    }
  }

  const processData = () => {
    let processed = [...allReports]
    
    // Search filter
    if (search) {
      const searchLower = search.toLowerCase()
      processed = processed.filter(item => 
        (item.product_name || '').toLowerCase().includes(searchLower) ||
        (item.category || '').toLowerCase().includes(searchLower) ||
        (item.product_code || '').toLowerCase().includes(searchLower)
      )
    }
    
    // ✅ Status filter menggunakan _status yang sudah dihitung
    if (statusFilter) {
      processed = processed.filter(item => {
        const status = item._status || getItemStatus(item)
        return status === statusFilter
      })
    }
    
    // Sort
    if (sortOrder === 'desc') {
      processed.sort((a, b) => (b.quantity || 0) - (a.quantity || 0))
    } else if (sortOrder === 'asc') {
      processed.sort((a, b) => (a.quantity || 0) - (b.quantity || 0))
    }
    
    setTotal(processed.length)
    setTotalPages(Math.ceil(processed.length / limit))
    
    const start = (page - 1) * limit
    const end = start + limit
    setReports(processed.slice(start, end))
  }

  // Helper function untuk mendapatkan status
  const getItemStatus = (item) => {
    const quantity = item.quantity || 0
    const minStock = item.min_stock || 0
    
    if (quantity === 0) return 'Out of Stock'
    if (quantity <= minStock) return 'Low Stock'
    return 'Healthy'
  }

  const handleExport = async () => {
    try {
      const loadingToast = toast.loading('Exporting report...')
      
      const response = await api.get('/reports/stock/export', {
        responseType: 'blob'
      })
      
      toast.dismiss(loadingToast)
      
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `stock_report_${new Date().toISOString().slice(0,10)}.xlsx`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
      
      toast.success('Report exported successfully')
    } catch (error) {
      toast.dismiss()
      toast.error('Failed to export report')
    }
  }

  const getStatusBadge = (status) => {
    switch (status) {
      case 'Out of Stock':
        return 'bg-red-100 text-red-800'
      case 'Low Stock':
        return 'bg-yellow-100 text-yellow-800'
      default:
        return 'bg-green-100 text-green-800'
    }
  }

  const getStatusText = (status) => {
    switch (status) {
      case 'Out of Stock':
        return 'Out of Stock'
      case 'Low Stock':
        return 'Low Stock'
      default:
        return 'Healthy'
    }
  }

  const toggleSort = () => {
    setSortOrder(sortOrder === 'desc' ? 'asc' : 'desc')
    setPage(1)
  }

  const getSortLabel = () => {
    if (sortOrder === 'desc') return 'Stock: Highest ↓'
    return 'Stock: Lowest ↑'
  }

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top',
        labels: {
          usePointStyle: true,
          padding: 20,
        }
      },
      title: {
        display: true,
        text: 'Top 20 Products Stock Overview',
        font: {
          size: 16,
          weight: 'bold'
        }
      }
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: {
          drawBorder: false,
        }
      },
      x: {
        grid: {
          display: false
        }
      }
    }
  }

  const pieOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'right',
        labels: {
          usePointStyle: true,
          padding: 20,
        }
      },
      title: {
        display: true,
        text: 'Stock Status Distribution',
        font: {
          size: 16,
          weight: 'bold'
        }
      }
    }
  }

  // ✅ Ambil status count dari summary yang sudah dihitung
  const statusCount = {
    healthy: summary?.healthy || 0,
    low: summary?.low_stock || 0,
    out: summary?.out_of_stock || 0
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Stock Reports</h2>
          {summary && (
            <div className="flex flex-wrap gap-4 mt-2 text-sm">
              <span className="text-gray-600">Total Products: <strong className="text-gray-900">{summary.total_products}</strong></span>
              <span className="text-gray-600">Total Stock: <strong className="text-gray-900">{summary.total_stock}</strong></span>
              <span className="text-yellow-600">⚠️ Low Stock: <strong>{summary.low_stock}</strong></span>
              <span className="text-red-600">🚫 Out of Stock: <strong>{summary.out_of_stock}</strong></span>
              <span className="text-green-600">✅ Healthy: <strong>{summary.healthy}</strong></span>
            </div>
          )}
        </div>
        <button
          onClick={handleExport}
          className="btn-secondary flex items-center gap-2"
        >
          <ArrowDownTrayIcon className="h-5 w-5" />
          Export Excel
        </button>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 card">
          {loading ? (
            <div className="flex justify-center items-center h-80">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
            </div>
          ) : chartData && reports.length > 0 ? (
            <div className="h-80">
              <Bar data={chartData} options={chartOptions} />
            </div>
          ) : (
            <p className="text-center text-gray-500 py-12">No data available</p>
          )}
        </div>
        <div className="card">
          {loading ? (
            <div className="flex justify-center items-center h-80">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
            </div>
          ) : pieChartData ? (
            <div className="h-80">
              <Pie data={pieChartData} options={pieOptions} />
            </div>
          ) : (
            <p className="text-center text-gray-500 py-12">No data available</p>
          )}
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search products by name, code, or category..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setPage(1)
            }}
            className="input-field pl-10"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value)
            setPage(1)
          }}
          className="input-field sm:w-48"
        >
          <option value="">All Status</option>
          <option value="Healthy">✅ Healthy</option>
          <option value="Low Stock">⚠️ Low Stock</option>
          <option value="Out of Stock">❌ Out of Stock</option>
        </select>
        <button
          onClick={toggleSort}
          className={`px-4 py-2 rounded-lg border transition-colors flex items-center gap-2 whitespace-nowrap ${
            sortOrder ? 'bg-blue-50 border-blue-500 text-blue-700' : 'border-gray-300 hover:bg-gray-50'
          }`}
        >
          <ArrowsUpDownIcon className="h-5 w-5" />
          {getSortLabel()}
          <span className="text-xs text-gray-500 ml-1">(Global)</span>
        </button>
      </div>

      {/* Summary Cards - Menggunakan data yang sudah dihitung */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-green-50 border border-green-200 rounded-xl p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-green-700 font-medium">Healthy Stock</p>
              <p className="text-2xl font-bold text-green-800">{statusCount.healthy}</p>
            </div>
            <div className="w-12 h-12 bg-green-200 rounded-full flex items-center justify-center">
              <span className="text-2xl">✅</span>
            </div>
          </div>
        </div>
        <div className="bg-yellow-50 border border-yellow-200 rounded-xl p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-yellow-700 font-medium">Low Stock</p>
              <p className="text-2xl font-bold text-yellow-800">{statusCount.low}</p>
            </div>
            <div className="w-12 h-12 bg-yellow-200 rounded-full flex items-center justify-center">
              <span className="text-2xl">⚠️</span>
            </div>
          </div>
        </div>
        <div className="bg-red-50 border border-red-200 rounded-xl p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-red-700 font-medium">Out of Stock</p>
              <p className="text-2xl font-bold text-red-800">{statusCount.out}</p>
            </div>
            <div className="w-12 h-12 bg-red-200 rounded-full flex items-center justify-center">
              <span className="text-2xl">🚫</span>
            </div>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr>
                <th className="table-header">Product</th>
                <th className="table-header">Category</th>
                <th className="table-header">Unit</th>
                <th className="table-header">Current Stock</th>
                <th className="table-header">Total In</th>
                <th className="table-header">Total Out</th>
                <th className="table-header">Min Stock</th>
                <th className="table-header">Status</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan="8" className="text-center py-8">
                    <div className="flex justify-center">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
                    </div>
                  </td>
                </tr>
              ) : reports.length === 0 ? (
                <tr>
                  <td colSpan="8" className="text-center py-8 text-gray-500">
                    No data found
                  </td>
                </tr>
              ) : (
                reports.map((item, index) => {
                  const status = item._status || getItemStatus(item)
                  return (
                    <tr key={index} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
                      <td className="table-cell font-medium">{item.product_name || '-'}</td>
                      <td className="table-cell">{item.category || '-'}</td>
                      <td className="table-cell">{item.unit || '-'}</td>
                      <td className="table-cell font-semibold text-blue-600">{item.quantity || 0}</td>
                      <td className="table-cell text-green-600">+{item.total_in || 0}</td>
                      <td className="table-cell text-red-600">-{item.total_out || 0}</td>
                      <td className="table-cell">{item.min_stock || 0}</td>
                      <td className="table-cell">
                        <span className={`px-2 py-1 text-xs font-semibold rounded-full ${getStatusBadge(status)}`}>
                          {getStatusText(status)}
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
              {search && <span className="text-blue-600 ml-2">(Filtered)</span>}
              {sortOrder && <span className="text-blue-600 ml-2">(Sorted globally)</span>}
            </div>
            
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage(1)}
                disabled={page === 1}
                className="px-3 py-1 rounded-lg border border-gray-300 disabled:opacity-50 hover:bg-gray-50 transition-colors"
              >
                <ChevronDoubleLeftIcon className="h-4 w-4" />
              </button>
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 rounded-lg border border-gray-300 disabled:opacity-50 hover:bg-gray-50 transition-colors"
              >
                <ChevronLeftIcon className="h-4 w-4" />
              </button>

              <div className="flex gap-1">
                {(() => {
                  const pages = []
                  const maxVisible = 7
                  const halfVisible = Math.floor(maxVisible / 2)
                  
                  let startPage = Math.max(1, page - halfVisible)
                  let endPage = Math.min(totalPages, page + halfVisible)
                  
                  if (page <= halfVisible + 1) {
                    endPage = Math.min(totalPages, maxVisible)
                  }
                  if (page > totalPages - halfVisible) {
                    startPage = Math.max(1, totalPages - maxVisible + 1)
                  }

                  if (startPage > 1) {
                    pages.push(1)
                    if (startPage > 2) pages.push('ellipsis')
                  }

                  for (let i = startPage; i <= endPage; i++) {
                    pages.push(i)
                  }

                  if (endPage < totalPages) {
                    if (endPage < totalPages - 1) pages.push('ellipsis')
                    pages.push(totalPages)
                  }

                  return pages.map((p, index) => {
                    if (p === 'ellipsis') {
                      return (
                        <span key={`ellipsis-${index}`} className="px-3 py-1 text-gray-500 select-none">
                          …
                        </span>
                      )
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
                    )
                  })
                })()}
              </div>

              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-3 py-1 rounded-lg border border-gray-300 disabled:opacity-50 hover:bg-gray-50 transition-colors"
              >
                <ChevronRightIcon className="h-4 w-4" />
              </button>
              <button
                onClick={() => setPage(totalPages)}
                disabled={page === totalPages}
                className="px-3 py-1 rounded-lg border border-gray-300 disabled:opacity-50 hover:bg-gray-50 transition-colors"
              >
                <ChevronDoubleRightIcon className="h-4 w-4" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default Reports