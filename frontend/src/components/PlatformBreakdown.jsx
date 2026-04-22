export default function PlatformBreakdown({ data }) {
  const colors = {
    android: '#3b82f6',
    ios: '#000000',
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-xl font-bold text-gray-800 mb-6">By Platform</h2>
      <div className="space-y-4">
        {data && data.length > 0 ? (
          data.map((platform) => (
            <div key={platform.platform}>
              <div className="flex justify-between items-center mb-1">
                <span className="text-gray-700 font-medium capitalize">
                  {platform.platform}
                </span>
                <span className="text-gray-600 text-sm">
                  {platform.total_devices} devices
                </span>
              </div>
              <div className="h-8 bg-gray-200 rounded-lg overflow-hidden">
                <div
                  className="h-full rounded-lg transition-all flex items-center justify-end pr-3"
                  style={{
                    width: `${Math.min(100, (platform.total_devices / Math.max(...data.map(p => p.total_devices || 1))) * 100)}%`,
                    backgroundColor: colors[platform.platform] || '#6b7280',
                  }}
                >
                  <span className="text-white text-xs font-semibold">
                    {(platform.delivery_rate * 100).toFixed(0)}%
                  </span>
                </div>
              </div>
            </div>
          ))
        ) : (
          <p className="text-gray-500 text-center py-8">No platform data</p>
        )}
      </div>
    </div>
  )
}
