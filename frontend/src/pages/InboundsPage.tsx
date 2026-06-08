import { useEffect, useState } from 'react'
import { apiGetInbounds, apiCreateInbound, apiDeleteInbound } from '../lib/api'

const protocols = ['vless', 'shadowsocks', 'trojan', 'hysteria2']
const transports = ['tcp', 'ws', 'grpc', 'h2', 'quic']

export function InboundsPage() {
  const [inbounds, setInbounds] = useState<any[]>([])
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ node_id: 1, protocol: 'vless', transport: 'tcp', port: 443, settings: '{}' })

  const load = () => apiGetInbounds().then(setInbounds)
  useEffect(() => { load() }, [])

  const create = async (e: any) => {
    e.preventDefault()
    await apiCreateInbound(form)
    setShowForm(false); load()
  }

  const remove = async (id: number) => {
    if (!confirm('Delete inbound?')) return
    await apiDeleteInbound(id); load()
  }

  const toggle = async (id: number) => {
    const token = localStorage.getItem('token')
    await fetch('/api/inbounds/' + id + '/toggle', { method: 'PATCH', headers: { Authorization: 'Bearer ' + token } })
    load()
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Inbounds</h2>
        <button onClick={() => setShowForm(!showForm)} className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm">+ Add</button>
      </div>

      {showForm && (
        <form onSubmit={create} className="bg-white dark:bg-neutral-800 p-4 rounded-xl shadow-sm mb-4 border">
          <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
            <select value={form.protocol} onChange={e => setForm({...form, protocol: e.target.value})}
              className="border rounded-lg px-3 py-2 dark:bg-neutral-700">
              {protocols.map(p => <option key={p}>{p}</option>)}
            </select>
            <select value={form.transport} onChange={e => setForm({...form, transport: e.target.value})}
              className="border rounded-lg px-3 py-2 dark:bg-neutral-700">
              {transports.map(t => <option key={t}>{t}</option>)}
            </select>
            <input type="number" value={form.port} onChange={e => setForm({...form, port: +e.target.value})}
              className="border rounded-lg px-3 py-2 dark:bg-neutral-700" placeholder="Port" />
            <input value={form.settings} onChange={e => setForm({...form, settings: e.target.value})}
              className="border rounded-lg px-3 py-2 dark:bg-neutral-700" placeholder='{"fingerprint":"chrome"}' />
            <button type="submit" className="bg-green-600 text-white px-4 py-2 rounded-lg">Create</button>
          </div>
          <button type="button" onClick={() => setShowForm(false)} className="text-red-500 text-sm mt-2">Cancel</button>
        </form>
      )}

      <div className="bg-white dark:bg-neutral-800 rounded-xl shadow-sm border overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-neutral-700">
            <tr>
              <th className="text-left px-4 py-3">Port</th>
              <th className="text-left px-4 py-3">Protocol</th>
              <th className="text-left px-4 py-3">Transport</th>
              <th className="text-left px-4 py-3">Node</th>
              <th className="text-left px-4 py-3">Status</th>
              <th className="text-center px-4 py-3">On</th>
              <th className="text-right px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {inbounds.map(ib => (
              <tr key={ib.id} className="border-t dark:border-neutral-700 hover:bg-gray-50">
                <td className="px-4 py-3 font-mono">:{ib.port}</td>
                <td className="px-4 py-3 uppercase text-xs">{ib.protocol}</td>
                <td className="px-4 py-3 text-xs">{ib.transport}</td>
                <td className="px-4 py-3 text-xs text-gray-500">{ib.Node?.name || '-'}</td>
                <td className="px-4 py-3">
                  <span className={"px-2 py-0.5 rounded text-xs " + (ib.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500')}>
                    {ib.is_active ? 'Active' : 'Off'}
                  </span>
                </td>
                <td className="px-4 py-3 text-center">
                  <button onClick={() => toggle(ib.id)}
                    className={"w-10 h-5 rounded-full transition-colors relative " + (ib.is_active ? 'bg-blue-500' : 'bg-gray-300')}>
                    <span className={"absolute top-0.5 w-4 h-4 bg-white rounded-full transition-all " + (ib.is_active ? 'left-5' : 'left-0.5')} />
                  </button>
                </td>
                <td className="px-4 py-3 text-right">
                  <button onClick={() => remove(ib.id)} className="text-red-500 text-xs">Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
