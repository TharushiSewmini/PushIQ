export default function MetricsCard({ title, value, subtitle }) {
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <p className="text-gray-600 text-sm font-medium">{title}</p>
      <p className="text-3xl font-bold text-gray-800 mt-2">{value.toLocaleString()}</p>
      <p className="text-gray-500 text-xs mt-2">{subtitle}</p>
    </div>
  )
}
