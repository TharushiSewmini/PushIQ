export default function DeliveryChart({ data }) {
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-xl font-bold text-gray-800 mb-4">Hourly Trends</h2>
      <div className="h-64 flex items-end justify-center space-x-1">
        {data && data.length > 0 ? (
          data.map((point, idx) => {
            const max = Math.max(...data.map(d => d.value))
            const height = max > 0 ? (point.value / max) * 100 : 0
            return (
              <div
                key={idx}
                className="flex-1 bg-blue-600 rounded-t hover:bg-blue-700 transition"
                style={{ height: `${height || 5}%` }}
                title={`${point.value} notifications`}
              />
            )
          })
        ) : (
          <p className="text-gray-500">No data available</p>
        )}
      </div>
    </div>
  )
}
