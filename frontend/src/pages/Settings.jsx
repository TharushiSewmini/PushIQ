export default function Settings({ setApiKey }) {
  const handleLogout = () => {
    localStorage.removeItem('pushiq_api_key')
    setApiKey('')
  }

  return (
    <div className="p-8">
      <h1 className="text-3xl font-bold text-gray-800 mb-6">Settings</h1>
      <div className="bg-white rounded-lg shadow p-6 max-w-md">
        <h2 className="text-xl font-bold text-gray-800 mb-4">Account</h2>
        <button
          onClick={handleLogout}
          className="bg-red-600 hover:bg-red-700 text-white font-semibold py-2 px-4 rounded-lg transition"
        >
          Logout
        </button>
      </div>
    </div>
  )
}
