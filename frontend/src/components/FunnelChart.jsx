export default function FunnelChart({ data }) {
  if (!data) return null

  const steps = [
    { label: 'Registered', value: data.registered_devices },
    { label: 'Online', value: data.online_devices },
    { label: 'Sent', value: data.notifications_sent },
    { label: 'Delivered', value: data.delivered },
  ]

  const maxValue = Math.max(...steps.map(s => s.value || 0))

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-xl font-bold text-gray-800 mb-6">Delivery Funnel</h2>
      <div className="space-y-4">
        {steps.map((step, idx) => (
          <div key={idx}>
            <div className="flex justify-between items-center mb-1">
              <span className="text-gray-700 font-medium">{step.label}</span>
              <span className="text-gray-600 text-sm">{step.value.toLocaleString()}</span>
            </div>
            <div className="h-8 bg-gray-200 rounded-lg overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-blue-500 to-blue-600 rounded-lg transition-all"
                style={{ width: maxValue > 0 ? `${(step.value / maxValue) * 100}%` : '0%' }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
