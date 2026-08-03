import React from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import { AuthProvider } from './context/AuthContext'
import ProtectedRoute from './components/common/ProtectedRoute'
import Layout from './components/Layout/Layout'
import Login from './pages/Login/Login'
import Dashboard from './pages/Dashboard/Dashboard'
import Products from './pages/Products/Products'
import Stock from './pages/Stock/Stock'
import TransactionIn from './pages/Transactions/TransactionIn'
import TransactionOut from './pages/Transactions/TransactionOut'
import Reports from './pages/Reports/Reports'
import Users from './pages/Users/Users'
import AIChatBot from './components/Chat/AIChatBot'

function App() {
  return (
    <Router>
      <AuthProvider>
        <Toaster 
          position="top-right"
          toastOptions={{
            duration: 3000,
            style: {
              background: '#363636',
              color: '#fff',
            },
            success: {
              duration: 3000,
              iconTheme: {
                primary: '#4F46E5',
                secondary: '#fff',
              },
            },
            error: {
              duration: 4000,
              iconTheme: {
                primary: '#EF4444',
                secondary: '#fff',
              },
            },
          }}
        />
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="products" element={<Products />} />
            <Route path="stock" element={<Stock />} />
            <Route path="transactions/in" element={<TransactionIn />} />
            <Route path="transactions/out" element={<TransactionOut />} />
            <Route path="reports" element={<Reports />} />
            <Route path="users" element={
              <ProtectedRoute allowedRoles={['superadmin']}>
                <Users />
              </ProtectedRoute>
            } />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        
        {/* AI Chat Bot - Muncul di semua halaman */}
        <AIChatBot />
      </AuthProvider>
    </Router>
  )
}

export default App