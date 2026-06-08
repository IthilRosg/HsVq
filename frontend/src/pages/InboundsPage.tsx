import { useEffect, useState } from 'react'
import { apiGetInbounds, apiCreateInbound, apiDeleteInbound } from '../lib/api'
import { Modal } from '../components/Modal'

const protocols = ['vless', 'shadowsocks', 'trojan', 'hysteria2']
const transports = ['tcp', 'ws', 'grpc', 'h2', 'quic']

export function InboundsPage() {
  const [inbounds, setInbounds] = useState<any[]>([])
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState({ node_id: 1, protocol: 'vless', transport: 'tcp', port: 443, settings: '{}' })

  const load = () => apiGetInbounds().then(setInbounds)
  useEffect(() => { load() }, [])

  const create = async (e: any) => {
    e.preventDefault()
    await apiCreateInbound(form)
    setShowModal(false); load()
  }

  const remove = async (id: number) => {
    if (!confirm('Delete this inbound?')) return
    await apiDeleteInbound(id); load()
  }

  const toggle = async (id: number) => {
    const t = localStorage.getItem('token')
    await fetch('/api/inbounds/' + id + '/toggle', { method: 'PATCH', headers: { Authorization: 'Bearer ' + t } })
    load()
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold">Inbounds</h2>
          <p className="text-sm text-gray-500 mt-0.5">Manage server connections</p>
        </div>
        <button onClick={() => setShowModal(true)}
          className="bg-blue-600 hover:bg-blue-700 active:scale-95 text-white px-4 py-2 rounded-lg text-sm font-medium transition-all duration-150 shadow-sm">
          + New Inbound
        </button>
      </div>

      <Modal open={showModal} onClose={() => setShowModal(false)} title="New Inbound">
        <form onSubmit={create} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-medium text-gray-500">Protocol</label>
              <select value={form.protocol} onChange={e => setForm({...form, protocol: e.target.value})}
                className="w-full mt-1 border rounded-lg px-3 py-2 text-sm bg-white dark:bg-neutral-700 focus:ring-2 focus:ring-blue-200 focus:border-blue-400 outline-none">
                {protocols.map(p => <option key={p}>{p}</option>)}
              </select>
            </div>
            <div>
              <label className="text-xs font-medium text-gray-500">Transport</label>
              <select value={form.transport} onChange={e => setForm({...form, transport: e.target.value})}
                className="w-full mt-1 border rounded-lg px-3 py-2 text-sm bg-white dark:bg-neutral-700 focus:ring-2 focus:ring-blue-200 outline-none">
                {transports.map(t => <option key={t}>{t}</option>)}
              </select>
            </div>
            <div>
              <label className="text-xs font-medium text-gray-500">Port</label>
              <input type="number" value={form.port} onChange={e => setForm({...form, port: +e.target.value})}
                className="w-full mt-1 border rounded-lg px-3 py-2 text-sm bg-white dark:bg-neutral-700 focus:ring-2 focus:ring-blue-200 outline-none" />
            </div>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={() => setShowModal(false)}
              className="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-all">Cancel</button>
            <button type="submit"
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 active:scale-95 rounded-lg transition-all shadow-sm">Create</button>
          </div>
        </form>
      </Modal>

      <div className="bg-white rounded-xl shadow-sm border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-gray-50/50">
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Port</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Protocol</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Transport</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Node</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="text-center px-4 py-3 text-xs font-medium text-gray-500 uppercase">On</th>
              <th className="text-right px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {inbounds.map(ib => (
              <tr key={ib.id} className="border-b last:border-0 hover:bg-gray-50/50 transition-colors">
                <td className="px-4 py-3 font-mono text-sm">:{ib.port}</td>
                <td className="px-4 py-3"><span className="text-xs font-mono bg-gray-100 px-2 py-0.5 rounded">{ib.protocol}</span></td>
                <td className="px-4 py-3 text-xs text-gray-600">{ib.transport}</td>
                <td className="px-4 py-3 text-xs text-gray-500">{ib.Node?.name || '-'}</td>
                <td className="px-4 py-3">
                  <span className={"inline-block w-2 h-2 rounded-full " + (ib.is_active ? 'bg-green-500' : 'bg-gray-300')} />
                </td>
                <td className="px-4 py-3 text-center">
                  <button onClick={() => toggle(ib.id)}
                    className={"relative inline-flex h-5 w-9 rounded-full transition-colors duration-200 " + (ib.is_active ? 'bg-blue-500' : 'bg-gray-200')}>
                    <span className={"inline-block h-4 w-4 transform rounded-full bg-white transition-transform duration-200 mt-0.5 " + (ib.is_active ? 'translate-x-[18px]' : 'translate-x-[2px]')} />
                  </button>
                </td>
                <td className="px-4 py-3 text-right">
                  <button onClick={() => remove(ib.id)} className="text-xs text-red-400 hover:text-red-600 transition-colors">Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
