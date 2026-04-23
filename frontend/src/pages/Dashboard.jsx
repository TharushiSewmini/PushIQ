import { useState, useEffect } from 'react'
import MetricsCard from '../components/MetricsCard'
import DeliveryChart from '../components/DeliveryChart'
import FunnelChart from '../components/FunnelChart'
import PlatformBreakdown from '../components/PlatformBreakdown'
import { apiUrl } from '../lib/api'

export default function Dashboard({ apiKey }) {
  const [dashboard, setDashboard] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    fetchDashboard()
  }, [apiKey])

  const fetchDashboard = async () => {
    try {
      setLoading(true)
      const response = await fetch(apiUrl('/api/v1/analytics/dashboard'), {
        headers: { 'X-Api-Key': apiKey }
      })
      
      if (!response.ok) throw new Error('Failed to fetch dashboard')
      const data = await response.json()
      setDashboard(data)
      setError(null)
    } catch (err) {
      setError(err.message)
      setDashboard(null)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="p-8 text-center text-gray-500">Loading dashboard...</div>
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
          {error}
          <button
            onClick={fetchDashboard}
            className="ml-4 underline hover:no-underline"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  if (!dashboard) return null

  return (
    <div className="p-8 space-y-8">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-800">Analytics Dashboard</h1>
        <button
          onClick={fetchDashboard}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
        >
          Refresh
        </button>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <MetricsCard
          title="Total Sent"
          value={dashboard.metrics?.total_sent || 0}
          subtitle="All time"
        />
        <MetricsCard
          title="Delivered"
          value={dashboard.metrics?.total_delivered || 0}
          subtitle={`${(dashboard.metrics?.delivery_rate * 100).toFixed(1)}% success`}
        />
        <MetricsCard
          title="Failed"
          value={dashboard.metrics?.total_failed || 0}
          subtitle={`${(dashboard.metrics?.failure_rate * 100).toFixed(1)}% failure`}
        />
        <MetricsCard
          title="Online Devices"
          value={dashboard.funnel_data?.online_devices || 0}
          subtitle={`${(dashboard.funnel_data?.online_rate * 100).toFixed(1)}% online`}
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <DeliveryChart data={dashboard?.hourly_trends || []} />
        <PlatformBreakdown data={dashboard?.by_platform || []} />
      </div>

      {/* Funnel and Top Notifications */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <FunnelChart data={dashboard?.funnel_data} />
        <TopNotificationsCard data={dashboard?.top_notifications || []} />
      </div>
    </div>
  )
}

function TopNotificationsCard({ data }) {
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-xl font-bold text-gray-800 mb-4">Top Notifications</h2>
      <div className="space-y-3">
        {data.slice(0, 5).map((notif) => (
          <div key={notif.notification_id} className="flex items-between text-sm">
            <span className="flex-1 text-gray-700 truncate">{notif.title}</span>
            <span className="text-gray-600">{notif.delivered}/{notif.sent}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
