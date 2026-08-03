import React, { useState, useEffect } from 'react'
import api from '../../api'
import { useAuth } from '../../context/AuthContext'
import { CubeIcon, ClipboardDocumentListIcon, ArrowDownTrayIcon, ArrowUpTrayIcon } from '@heroicons/react/24/outline'

const Dashboard = () => {
  const { user } = useAuth()
  const [stats, setStats] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchDashboardData()
  }, [])

  const fetchDashboardData = async () => {
    try {
      const response = await api.get('/reports/dashboard')
      setStats(response.data.data)
    } catch (error) {
      console.error('Failed to load dashboard data')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  const statCards = [
    { title: 'Total Products', value: stats?.total_products || 0, icon: CubeIcon, color: 'bg-blue-500' },
    { title: 'Total Stock', value: stats?.total_stock || 0, icon: ClipboardDocumentListIcon, color: 'bg-green-500' },
    { title: 'Total In', value: stats?.total_in || 0, icon: ArrowDownTrayIcon, color: 'bg-purple-500' },
    { title: 'Total Out', value: stats?.total_out || 0, icon: ArrowUpTrayIcon, color: 'bg-orange-500' }
  ]

  return (
    <div className="space-y-6">
      <div className="bg-gradient-to-r from-primary-500 to-primary-600 rounded-xl p-6 text-white">
        <h2 className="text-2xl font-bold">Welcome back, {user?.fullname}! 👋</h2>
        <p className="mt-1 text-primary-100">Here's what's happening with your inventory today</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => (
          <div key={index} className="bg-white rounded-xl shadow-lg p-6 hover:shadow-xl transition-shadow duration-300">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500">{stat.title}</p>
                <p className="text-2xl font-bold mt-1">{stat.value}</p>
              </div>
              <div className={`p-3 rounded-xl ${stat.color} bg-opacity-10`}>
                <stat.icon className={`h-6 w-6 ${stat.color.replace('bg-', 'text-')}`} />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default Dashboard