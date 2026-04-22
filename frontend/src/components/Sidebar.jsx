import { Link, useLocation } from 'react-router-dom'

export default function Sidebar({ open }) {
  const location = useLocation()
  
  const menuItems = [
    { path: '/', label: 'Dashboard', icon: 'LayoutDashboard' },
    { path: '/notifications', label: 'Notifications', icon: 'Bell' },
    { path: '/analytics', label: 'Analytics', icon: 'BarChart3' },
    { path: '/settings', label: 'Settings', icon: 'Settings' },
  ]

  return (
    <aside className={`
      w-64 bg-gray-900 text-white transition-transform duration-300
      ${open ? 'translate-x-0' : '-translate-x-full'}
      md:translate-x-0 fixed md:relative h-screen flex flex-col z-50
    `}>
      <div className="p-6 border-b border-gray-700">
        <h2 className="text-xl font-bold">PushIQ</h2>
      </div>
      
      <nav className="flex-1 p-6 space-y-2">
        {menuItems.map((item) => (
          <Link
            key={item.path}
            to={item.path}
            className={`
              block px-4 py-2 rounded-lg transition
              ${location.pathname === item.path
                ? 'bg-blue-600 text-white'
                : 'text-gray-300 hover:bg-gray-800'
              }
            `}
          >
            {item.label}
          </Link>
        ))}
      </nav>

      <div className="p-6 border-t border-gray-700">
        <div className="text-xs text-gray-400">v1.0.0</div>
      </div>
    </aside>
  )
}
