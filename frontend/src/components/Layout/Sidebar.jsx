import React from 'react'
import { NavLink } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import {
  HomeIcon,
  CubeIcon,
  ClipboardDocumentListIcon,
  ArrowDownTrayIcon,
  ArrowUpTrayIcon,
  ChartBarIcon,
  UsersIcon,
  XMarkIcon
} from '@heroicons/react/24/outline'

const Sidebar = ({ isOpen, setIsOpen }) => {
  const { user } = useAuth()

  const menuItems = [
    { path: '/dashboard', icon: HomeIcon, label: 'Dashboard' },
    { path: '/products', icon: CubeIcon, label: 'Products' },
    { path: '/stock', icon: ClipboardDocumentListIcon, label: 'Stock' },
    { path: '/transactions/in', icon: ArrowDownTrayIcon, label: 'Transaction In' },
    { path: '/transactions/out', icon: ArrowUpTrayIcon, label: 'Transaction Out' },
    { path: '/reports', icon: ChartBarIcon, label: 'Reports' },
  ]

  if (user?.role === 'superadmin') {
    menuItems.push({ path: '/users', icon: UsersIcon, label: 'Users' })
  }

  return (
    <>
      {isOpen && (
        <div 
          className="fixed inset-0 z-20 bg-black bg-opacity-50 lg:hidden"
          onClick={() => setIsOpen(false)}
        />
      )}

      <aside 
        className={`
          fixed lg:static inset-y-0 left-0 z-30
          w-64 bg-white border-r border-gray-200
          transform transition-transform duration-300 ease-in-out
          ${isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        `}
      >
        <div className="flex items-center justify-between h-16 px-4 border-b border-gray-200">
          <div className="flex items-center gap-2">
            <CubeIcon className="h-8 w-8 text-primary-600" />
            <span className="text-xl font-bold text-primary-600">Inventory</span>
          </div>
          <button 
            className="lg:hidden p-2 rounded-lg hover:bg-gray-100"
            onClick={() => setIsOpen(false)}
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        <nav className="p-4 space-y-1">
          {menuItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) => 
                `flex items-center gap-3 px-4 py-3 rounded-lg transition-all duration-200 ${
                  isActive 
                    ? 'bg-primary-50 text-primary-600 font-semibold' 
                    : 'text-gray-700 hover:bg-primary-50 hover:text-primary-600'
                }`
              }
              onClick={() => setIsOpen(false)}
            >
              <item.icon className="h-5 w-5" />
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>

        <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-200">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-primary-100 flex items-center justify-center">
              <span className="text-primary-600 font-semibold">
                {user?.fullname?.charAt(0) || 'U'}
              </span>
            </div>
            <div>
              <p className="text-sm font-medium">{user?.fullname}</p>
              <p className="text-xs text-gray-500 capitalize">{user?.role}</p>
            </div>
          </div>
        </div>
      </aside>
    </>
  )
}

export default Sidebar