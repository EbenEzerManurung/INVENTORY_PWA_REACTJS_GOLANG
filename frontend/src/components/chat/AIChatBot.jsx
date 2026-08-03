import React, { useState, useEffect, useRef } from 'react'
import api from '../../api'
import { useAuth } from '../../context/AuthContext'
import toast from 'react-hot-toast'
import {
  ChatBubbleLeftRightIcon,
  XMarkIcon,
  PaperAirplaneIcon,
  UserCircleIcon,
  SparklesIcon,
  TrashIcon,
  ChevronDownIcon
} from '@heroicons/react/24/outline'

const AIChatBot = () => {
  const { user } = useAuth()
  const [isOpen, setIsOpen] = useState(false)
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [showScrollButton, setShowScrollButton] = useState(false)
  const messagesEndRef = useRef(null)
  const chatContainerRef = useRef(null)
  const inputRef = useRef(null)

  // ==================== GET GREETING BASED ON TIME ====================
  const getGreeting = () => {
    const hour = new Date().getHours()
    let timeGreeting = ''
    let emoji = ''
    
    if (hour >= 5 && hour < 12) {
      timeGreeting = 'Good Morning'
      emoji = '🌅'
    } else if (hour >= 12 && hour < 17) {
      timeGreeting = 'Good Afternoon'
      emoji = '☀️'
    } else if (hour >= 17 && hour < 20) {
      timeGreeting = 'Good Evening'
      emoji = '🌤️'
    } else if (hour >= 20 && hour < 24) {
      timeGreeting = 'Good Evening'
      emoji = '🌙'
    } else {
      timeGreeting = 'Good Night'
      emoji = '🌃'
    }
    
    const userName = user?.fullname || user?.username || 'Guest'
    
    return `👋 **Halo ${userName}, ${timeGreeting}!** ${emoji}\n\n` +
           `📊 **I can help you with:**\n` +
           `  • 📦 Stock reports & summaries\n` +
           `  • 📋 Transaction history (In/Out)\n` +
           `  • 🔔 Pending approvals\n` +
           `  • 👤 User & role information\n` +
           `  • ⚠️ Low stock alerts\n\n` +
           `💡 _What would you like to know today?_`
  }

  // ==================== SCROLL FUNCTIONS ====================
  const scrollToBottom = () => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ 
        behavior: 'smooth', 
        block: 'end' 
      })
    }
  }

  // ==================== GREETING - MUNCUL SAAT CHAT DIBUKA ====================
  useEffect(() => {
    if (isOpen && messages.length === 0) {
      const newGreeting = getGreeting()
      setMessages([
        {
          id: 1,
          text: newGreeting,
          sender: 'bot',
          timestamp: new Date()
        }
      ])
      setTimeout(scrollToBottom, 300)
    }
  }, [isOpen])

  // ==================== UPDATE GREETING SAAT USER BERUBAH ====================
  useEffect(() => {
    if (user && isOpen && messages.length > 0 && messages[0].sender === 'bot') {
      const newGreeting = getGreeting()
      setMessages(prev => {
        const updated = [...prev]
        updated[0] = { ...updated[0], text: newGreeting }
        return updated
      })
    }
  }, [user])

  // Check scroll position
  const handleScroll = () => {
    if (chatContainerRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = chatContainerRef.current
      const isAtBottom = scrollHeight - scrollTop - clientHeight < 100
      setShowScrollButton(!isAtBottom && messages.length > 3)
    }
  }

  // Focus input when chat opens
  useEffect(() => {
    if (isOpen && inputRef.current) {
      setTimeout(() => inputRef.current.focus(), 500)
    }
  }, [isOpen])

  // ==================== SEND MESSAGE ====================
  const handleSend = async () => {
    const message = input.trim()
    if (!message) return

    const userMessage = {
      id: Date.now(),
      text: message,
      sender: 'user',
      timestamp: new Date()
    }
    setMessages(prev => [...prev, userMessage])
    setInput('')
    setLoading(true)

    setTimeout(scrollToBottom, 100)

    try {
      const response = await api.post('/ai/chat', { message })
      const botMessage = {
        id: Date.now() + 1,
        text: response.data.data.response || 'Sorry, I could not process your request.',
        sender: 'bot',
        timestamp: new Date()
      }
      setMessages(prev => [...prev, botMessage])
    } catch (error) {
      console.error('Chat error:', error)
      setMessages(prev => [...prev, {
        id: Date.now() + 1,
        text: '❌ Sorry, something went wrong. Please try again.',
        sender: 'bot',
        timestamp: new Date()
      }])
    } finally {
      setLoading(false)
    }
  }

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const clearChat = () => {
    setMessages([])
  }

  const formatTime = (date) => {
    return new Date(date).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: true
    })
  }

  const suggestions = [
    'Show products with highest stock',
    'What transactions are pending approval?',
    'Show stock summary',
    'Who are the Head users?',
    'Show low stock products'
  ]

  return (
    <>
      {/* Chat Button - HANYA SATU */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="fixed bottom-6 right-6 z-50 p-4 bg-gradient-to-r from-indigo-500 to-purple-600 text-white rounded-full shadow-2xl hover:shadow-3xl transition-all duration-300 hover:scale-110 group"
        style={{ boxShadow: '0 8px 32px rgba(99, 102, 241, 0.4)' }}
      >
        {isOpen ? (
          <XMarkIcon className="h-6 w-6" />
        ) : (
          <div className="relative">
            <ChatBubbleLeftRightIcon className="h-6 w-6" />
            <span className="absolute -top-1 -right-1 w-3.5 h-3.5 bg-green-400 rounded-full border-2 border-white animate-pulse"></span>
          </div>
        )}
      </button>

      {/* Chat Window - HANYA SATU */}
      {isOpen && (
        <div className="fixed bottom-24 right-6 z-50 w-[480px] h-[650px] bg-white rounded-2xl shadow-2xl flex flex-col border border-gray-100 overflow-hidden transition-all duration-300">
          {/* Header */}
          <div className="bg-gradient-to-r from-indigo-600 to-purple-600 text-white p-4 flex items-center justify-between shrink-0">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-white/20 rounded-full">
                <SparklesIcon className="h-5 w-5" />
              </div>
              <div>
                <h3 className="font-semibold text-sm">AI Inventory Assistant</h3>
                <p className="text-xs text-purple-200 flex items-center gap-1">
                  <span className="w-2 h-2 bg-green-400 rounded-full inline-block"></span>
                  Online • Ready to help
                </p>
              </div>
            </div>
            <div className="flex gap-1">
              <button
                onClick={clearChat}
                className="p-1.5 hover:bg-white/20 rounded-lg transition-colors"
                title="Clear chat"
              >
                <TrashIcon className="h-4 w-4" />
              </button>
              <button
                onClick={() => setIsOpen(false)}
                className="p-1.5 hover:bg-white/20 rounded-lg transition-colors"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>

          {/* Messages */}
          <div 
            ref={chatContainerRef}
            onScroll={handleScroll}
            className="flex-1 overflow-y-auto pt-8 pb-4 px-6 space-y-4 bg-gray-50"
            style={{ scrollBehavior: 'smooth' }}
          >
            {messages.map((msg) => (
              <div
                key={msg.id}
                className={`flex ${msg.sender === 'user' ? 'justify-end' : 'justify-start'} animate-fadeIn`}
              >
                <div className={`max-w-[85%] ${msg.sender === 'user' ? 'order-2' : 'order-1'}`}>
                  <div className={`flex items-start gap-3 ${msg.sender === 'user' ? 'flex-row-reverse' : ''}`}>
                    {msg.sender === 'bot' ? (
                      <div className="w-9 h-9 rounded-full bg-gradient-to-r from-indigo-500 to-purple-600 flex items-center justify-center flex-shrink-0 shadow-md mt-0.5">
                        <SparklesIcon className="h-5 w-5 text-white" />
                      </div>
                    ) : (
                      <div className="w-9 h-9 rounded-full bg-gradient-to-r from-gray-400 to-gray-500 flex items-center justify-center flex-shrink-0 shadow-md mt-0.5">
                        <UserCircleIcon className="h-6 w-6 text-white" />
                      </div>
                    )}
                    <div>
                      <div
                        className={`rounded-2xl px-5 py-3.5 text-sm leading-relaxed whitespace-pre-wrap ${
                          msg.sender === 'user'
                            ? 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white shadow-md'
                            : 'bg-white shadow-md border border-gray-100'
                        }`}
                      >
                        {msg.text}
                      </div>
                      <div className={`text-[10px] text-gray-400 mt-1.5 ${msg.sender === 'user' ? 'text-right' : ''}`}>
                        {formatTime(msg.timestamp)}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))}
            {loading && (
              <div className="flex justify-start animate-fadeIn">
                <div className="flex items-start gap-3">
                  <div className="w-9 h-9 rounded-full bg-gradient-to-r from-indigo-500 to-purple-600 flex items-center justify-center flex-shrink-0 shadow-md mt-0.5">
                    <SparklesIcon className="h-5 w-5 text-white" />
                  </div>
                  <div className="bg-white shadow-md border border-gray-100 rounded-2xl px-5 py-3.5">
                    <div className="flex items-center gap-1">
                      <div className="w-2 h-2 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
                      <div className="w-2 h-2 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
                      <div className="w-2 h-2 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
                    </div>
                  </div>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Scroll to bottom button */}
          {showScrollButton && (
            <button
              onClick={scrollToBottom}
              className="absolute bottom-24 right-4 z-10 p-2 bg-white text-indigo-600 rounded-full shadow-lg border border-gray-200 hover:bg-gray-50 transition-all duration-200"
            >
              <ChevronDownIcon className="h-5 w-5" />
            </button>
          )}

          {/* Suggestions */}
          {messages.length <= 2 && (
            <div className="px-4 py-3 border-t border-gray-100 bg-gray-50 shrink-0">
              <p className="text-xs text-gray-500 mb-2 font-medium">✨ Quick questions:</p>
              <div className="flex flex-wrap gap-2">
                {suggestions.map((suggestion, index) => (
                  <button
                    key={index}
                    onClick={() => {
                      setInput(suggestion)
                      setTimeout(handleSend, 150)
                    }}
                    className="px-3 py-1.5 text-xs bg-white border border-gray-200 rounded-full hover:bg-indigo-50 hover:border-indigo-300 hover:text-indigo-600 transition-all duration-200 shadow-sm hover:shadow"
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Input */}
          <div className="p-4 border-t border-gray-200 bg-white shrink-0">
            <div className="flex gap-2">
              <input
                ref={inputRef}
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Ask me anything..."
                className="flex-1 px-4 py-2.5 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-sm transition-all duration-200 bg-gray-50"
                disabled={loading}
              />
              <button
                onClick={handleSend}
                disabled={loading || !input.trim()}
                className="p-2.5 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl hover:from-indigo-700 hover:to-purple-700 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-md hover:shadow-lg"
              >
                <PaperAirplaneIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* CSS for animations */}
      <style>{`
        @keyframes fadeIn {
          from { 
            opacity: 0; 
            transform: translateY(12px) scale(0.98);
          }
          to { 
            opacity: 1; 
            transform: translateY(0) scale(1);
          }
        }
        .animate-fadeIn {
          animation: fadeIn 0.35s ease-out forwards;
        }
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
        .animate-pulse {
          animation: pulse 2s ease-in-out infinite;
        }
        @keyframes bounce {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(-4px); }
        }
        .animate-bounce {
          animation: bounce 1s ease-in-out infinite;
        }
        .shadow-3xl {
          box-shadow: 0 12px 48px rgba(0, 0, 0, 0.2);
        }
      `}</style>
    </>
  )
}

export default AIChatBot