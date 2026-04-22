import { useState, useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Header from './components/Header'
import Sidebar from './components/Sidebar'
import Dashboard from './pages/Dashboard'
import Notifications from './pages/Notifications'
import Analytics from './pages/Analytics'
import Settings from './pages/Settings'

export default function App() {
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [apiKey, setApiKey] = useState(localStorage.getItem('pushiq_api_key') || '')

  useEffect(() => {
    if (apiKey) {
      localStorage.setItem('pushiq_api_key', apiKey)
    }
  }, [apiKey])

  if (!apiKey) {
    return <LoginPage setApiKey={setApiKey} />
  }

  return (
    <Router>
      <div className="flex h-screen bg-gray-100">
        <Sidebar open={sidebarOpen} />
        <div className="flex-1 flex flex-col">
          <Header onMenuClick={() => setSidebarOpen(!sidebarOpen)} />
          <main className="flex-1 overflow-auto">
            <Routes>
              <Route path="/" element={<Dashboard apiKey={apiKey} />} />
              <Route path="/notifications" element={<Notifications apiKey={apiKey} />} />
              <Route path="/analytics" element={<Analytics apiKey={apiKey} />} />
              <Route path="/settings" element={<Settings setApiKey={setApiKey} />} />
            </Routes>
          </main>
        </div>
      </div>
    </Router>
  )
}

function LoginPage({ setApiKey }) {
  const [key, setKey] = useState('')
  const [error, setError] = useState('')

  const handleLogin = async () => {
    if (!key.trim()) {
      setError('API Key is required')
      return
    }

    try {
      const response = await fetch('http://localhost:8080/health', {
        headers: { 'X-Api-Key': key }
      })
      
      if (response.ok) {
        setApiKey(key)
      } else {
        setError('Invalid API Key')
      }
    } catch (err) {
      setError('Unable to connect to PushIQ API')
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center">
      <div className="bg-white rounded-lg shadow-lg p-8 w-96">
        <h1 className="text-3xl font-bold text-gray-800 mb-2">PushIQ</h1>
        <p className="text-gray-600 mb-8">Analytics Dashboard</p>
        
        <div className="space-y-4">
          <input
            type="password"
            placeholder="Enter API Key"
            value={key}
            onChange={(e) => {
              setKey(e.target.value)
              setError('')
            }}
            onKeyPress={(e) => e.key === 'Enter' && handleLogin()}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-600"
          />
          
          {error && <div className="text-red-600 text-sm">{error}</div>}
          
          <button
            onClick={handleLogin}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-4 rounded-lg transition"
          >
            Login
          </button>
        </div>
      </div>
    </div>
  )
}
