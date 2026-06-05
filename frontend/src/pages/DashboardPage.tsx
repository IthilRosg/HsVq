import { useEffect, useState } from 'react'
import { apiGetDashboard } from '../lib/api'

export function DashboardPage() {
  const [data, setData] = useState(null)
  useEffect(() => { apiGetDashboard().then(setData).catch(console.error) }, [])
  if (!data) return <div className="text-center py-12 text-gray-500">Loading...</div>
  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">Dashboard</h2>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        <div className="bg-white dark:bg-neutral-800 p-6 rounded-xl shadow-sm border">
          <div className="text-sm text-gray-500">Users</div>
          <div className="text-3xl font-bold mt-1">{data.total_users}</div>
        </div>
        <div className="bg-white dark:bg-neutral-800 p-6 rounded-xl shadow-sm border">
          <div className="text-sm text-gray-500">Active Subs</div>
          <div className="text-3xl font-bold mt-1">{data.active_subs}</div>
        </div>
        <div className="bg-white dark:bg-neutral-800 p-6 rounded-xl shadow-sm border">
          <div className="text-sm text-gray-500">Main Node</div>
          <div className="text-lg font-bold mt-1">{data.main_node?.name || '-'}</div>
          <div className="text-xs text-gray-400">{data.main_node?.address}</div>
        </div>
      </div>
    </div>
  )
}
