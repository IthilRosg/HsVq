import { useEffect, useState } from 'react'
import { apiGetDashboard } from '../lib/api'

function BarChart({ data, labelKey, valueKey, color }: { data: any[], labelKey: string, valueKey: string, color: string }) {
  const max = Math.max(...data.map(d => d[valueKey]), 1)
  return (
    <div className="flex items-end gap-1 h-24">
      {data.map((d, i) => (
        <div key={i} className="flex-1 flex flex-col items-center gap-0.5">
          <div className="text-[8px] text-gray-400">{(d[valueKey] / 1e9).toFixed(1)}</div>
          <div style={{ height: Math.max((d[valueKey] / max) * 80, 2) + 'px' }}
            className={"w-full rounded-t " + color} />
          <div className="text-[8px] text-gray-400">{d[labelKey].slice(5)}</div>
        </div>
      ))}
    </div>
  )
}

export function DashboardPage() {
  const [data, setData] = useState<any>(null)
  const [history, setHistory] = useState<any[]>([])

  useEffect(() => {
    apiGetDashboard().then(setData).catch(console.error)
    const token = localStorage.getItem('token')
    fetch('/api/traffic/history?days=7', { headers: { Authorization: 'Bearer ' + token } })
      .then(r => r.json()).then(setHistory).catch(console.error)
  }, [])

  if (!data) return <div className="text-center py-12 text-gray-500">Loading...</div>

  const fmt = (gb: number) => gb >= 1000 ? (gb/1000).toFixed(1) + ' TB' : gb.toFixed(1) + ' GB'

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">Dashboard</h2>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white dark:bg-neutral-800 p-5 rounded-xl shadow-sm border">
          <div className="text-xs text-gray-500">Users</div>
          <div className="text-2xl font-bold mt-1">{data.total_users}</div>
        </div>
        <div className="bg-white dark:bg-neutral-800 p-5 rounded-xl shadow-sm border">
          <div className="text-xs text-gray-500">Active Subs</div>
          <div className="text-2xl font-bold mt-1">{data.active_subs}</div>
        </div>
        <div className="bg-white dark:bg-neutral-800 p-5 rounded-xl shadow-sm border">
          <div className="text-xs text-gray-500">Expiring Soon</div>
          <div className="text-2xl font-bold mt-1">{data.expiring_soon||0}</div>
        </div>
        <div className="bg-white dark:bg-neutral-800 p-5 rounded-xl shadow-sm border">
          <div className="text-xs text-gray-500">Today</div>
          <div className="text-2xl font-bold mt-1">{fmt(data.traffic_today?.down_gb||0)}</div>
        </div>
      </div>

      {history.length > 0 && (
        <div className="bg-white dark:bg-neutral-800 p-5 rounded-xl shadow-sm border mb-6">
          <h3 className="text-sm font-semibold mb-3">Traffic (7 days)</h3>
          <div className="flex gap-6">
            <div className="flex-1">
              <div className="text-[10px] text-gray-400 mb-1">Download</div>
              <BarChart data={history} labelKey="date" valueKey="down" color="bg-blue-400" />
            </div>
            <div className="flex-1">
              <div className="text-[10px] text-gray-400 mb-1">Upload</div>
              <BarChart data={history} labelKey="date" valueKey="up" color="bg-orange-400" />
            </div>
          </div>
        </div>
      )}

      <div className="bg-white dark:bg-neutral-800 p-5 rounded-xl shadow-sm border">
        <h3 className="text-sm font-semibold mb-2">Server</h3>
        <div className="text-xs text-gray-500">
          Node: {data.main_node?.name || '-'} ({data.main_node?.address})<br />
          Time: {new Date(data.server_time).toLocaleString('ru-RU')}
        </div>
      </div>
    </div>
  )
}
