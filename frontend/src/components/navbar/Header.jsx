import React from 'react'
import { useAuth } from '../../context/AuthContext'
import { Bars3Icon, CubeIcon } from '@heroicons/react/24/outline'
import ProfileDropdown from '../Navbar/ProfileDropdown'

const Header = ({ toggleSidebar }) => {
  const { user } = useAuth()

  return (
    <header className="bg-white border-b border-gray-200 h-16 px-6 flex items-center justify-between">
      <div className="flex items-center gap-4">
        <button
          onClick={toggleSidebar}
          className="p-2 rounded-lg hover:bg-gray-100 lg:hidden"
        >
          <Bars3Icon className="h-6 w-6" />
        </button>
        
        {/* Logo */}
        <div className="flex items-center gap-2">
          <div className="p-1.5 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg shadow-md shadow-blue-500/30">
            <CubeIcon className="h-6 w-6 text-white" />
          </div>
          <span className="text-lg font-bold text-blue-600 hidden sm:block">Inventory</span>
        </div>
        
        <h1 className="text-xl font-semibold text-gray-900 ml-2">
          {document.title || 'Dashboard'}
        </h1>
      </div>

      <div className="flex items-center gap-4">
        <ProfileDropdown />
      </div>
    </header>
  )
}

export default Header