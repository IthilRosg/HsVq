import { useEffect, useState } from 'react'

const api = (path: string, opts?: any) => {
  const t = localStorage.getItem('token')
  return fetch('/api' + path, {
    ...opts,
    headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + t, ...opts?.headers }
  }).then(r => r.json())
}

export function ProfilesPage() {
  const [profiles, setProfiles] = useState<any[]>([])
  const [plans, setPlans] = useState<any[]>([])
  const [inbounds, setInbounds] = useState<any[]>([])
  const [showForm, setShowForm] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [form, setForm] = useState({ name: '', plan_id: 1, node_id: 1, inbound_ids: [] as number[], expires_at: '' })

  const load = () => { api('/profiles').then(setProfiles); api('/plans').then(setPlans); api('/inbounds').then(setInbounds) }
  useEffect(() => { load() }, [])

  const create = async (e: any) => {
    e.preventDefault()
    await api('/profiles', { method: 'POST', body: JSON.stringify(form) })
    setShowForm(false); resetForm(); load()
  }

  const update = async (e: any) => {
    e.preventDefault()
    await api('/profiles/' + editId, { method: 'PUT', body: JSON.stringify(form) })
    setEditId(null); resetForm(); load()
  }

  const resetForm = () => setForm({ name: '', plan_id: 1, node_id: 1, inbound_ids: [], expires_at: '' })

  const remove = async (id: number) => {
    if (!confirm('Delete?')) return
    await api('/profiles/' + id, { method: 'DELETE' }); load()
  }

  const startEdit = (p: any) => {
    setForm({
      name: p.Name || p.name || '',
      plan_id: p.PlanID || p.plan_id || 1,
      node_id: p.NodeID || p.node_id || 1,
      inbound_ids: JSON.parse(p.InboundIDs || p.inbound_ids || '[]'),
      expires_at: p.ExpiresAt ? p.ExpiresAt.slice(0,10) : p.expires_at ? p.expires_at.slice(0,10) : ''
    })
    setEditId(p.id)
  }

  const toggleInbound = (id: number) => {
    setForm((f: any) => ({
      ...f, inbound_ids: f.inbound_ids.includes(id) ? f.inbound_ids.filter((i: number) => i !== id) : [...f.inbound_ids, id]
    }))
  }

  const subURL = (p: any) => p.SubscriptionURL || p.subscription_url || ''

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Clients</h2>
        <button onClick={() => { setShowForm(!showForm); setEditId(null); resetForm() }} className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm">+ Add Client</button>
      </div>

      {(showForm || editId) && (
        <form onSubmit={editId ? update : create} className="bg-white dark:bg-neutral-800 p-4 rounded-xl shadow-sm mb-4 border">
          <div className="grid grid-cols-2 gap-3 mb-3">
            <div>
              <label className="text-xs text-gray-500">Name *</label>
              <input value={form.name} onChange={e => setForm({...form, name: e.target.value})} required
                className="w-full border rounded-lg px-3 py-2 dark:bg-neutral-700" />
            </div>
            <div>
              <label className="text-xs text-gray-500">Plan</label>
              <select value={form.plan_id} onChange={e => setForm({...form, plan_id: +e.target.value})}
                className="w-full border rounded-lg px-3 py-2 dark:bg-neutral-700">
                {plans.map((p: any) => <option key={p.id} value={p.id}>{p.name}</option>)}
              </select>
            </div>
            <div>
              <label className="text-xs text-gray-500">Expiry</label>
              <input type="date" value={form.expires_at} onChange={e => setForm({...form, expires_at: e.target.value})}
                className="w-full border rounded-lg px-3 py-2 dark:bg-neutral-700" />
            </div>
          </div>
          <div className="mb-3">
            <label className="text-xs text-gray-500">Protocols</label>
            <div className="flex flex-wrap gap-2 mt-1">
              {inbounds.filter((ib: any) => ib.is_active).map((ib: any) => (
                <label key={ib.id}
                  className={"px-3 py-1.5 border rounded-lg cursor-pointer text-sm " + (form.inbound_ids.includes(ib.id) ? 'bg-blue-50 border-blue-400' : '')}>
                  <input type="checkbox" checked={form.inbound_ids.includes(ib.id)} onChange={() => toggleInbound(ib.id)} className="sr-only" />
                  {ib.protocol.toUpperCase()} + {ib.transport}
                </label>
              ))}
            </div>
          </div>
          <div className="flex gap-2">
            <button type="submit" className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm">{editId ? 'Update' : 'Create'}</button>
            <button type="button" onClick={() => { setShowForm(false); setEditId(null); resetForm() }} className="text-red-500 text-sm px-4 py-2">Cancel</button>
          </div>
        </form>
      )}

      <div className="bg-white dark:bg-neutral-800 rounded-xl shadow-sm border overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-neutral-700">
            <tr>
              <th className="text-left px-4 py-3">Name</th>
              <th className="text-left px-4 py-3">Plan</th>
              <th className="text-left px-4 py-3">Status</th>
              <th className="text-left px-4 py-3">Expires</th>
              <th className="text-left px-4 py-3">URL</th>
              <th className="text-right px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {profiles.map((p: any) => (
              <tr key={p.id} className="border-t dark:border-neutral-700 hover:bg-gray-50">
                <td className="px-4 py-3 font-medium">{p.Name || p.name}</td>
                <td className="px-4 py-3">{p.Plan?.name || '-'}</td>
                <td className="px-4 py-3">
                  <span className={"px-2 py-0.5 rounded text-xs " + ((p.Status || p.status) === 'active' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700')}>
                    {p.Status || p.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-xs">{new Date(p.ExpiresAt || p.expires_at).toLocaleDateString()}</td>
                <td className="px-4 py-3">
                  <div className="flex items-center gap-1">
                    <span className="text-xs text-gray-400 truncate max-w-[120px]">{subURL(p).slice(0, 30)}...</span>
                    <button onClick={() => navigator.clipboard.writeText(subURL(p))}
                      className="text-blue-600 text-xs hover:underline shrink-0">Copy</button>
                    <button onClick={() => window.open("/api/configs/user/" + p.id + "/qr")}
                      className="text-green-600 text-xs hover:underline shrink-0">QR</button>
                  </div>
                </td>
                <td className="px-4 py-3 text-right">
                  <button onClick={() => startEdit(p)} className="text-blue-500 text-xs mr-2">Edit</button>
                  <button onClick={() => remove(p.id)} className="text-red-500 text-xs">Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
