import { useEffect, useState } from 'react'

export function PlansPage() {
  const [plans, setPlans] = useState([])
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', duration_days: 30, traffic_limit: 500000000000, device_limit: 10, price: 500 })

  const load = () => {
    const token = localStorage.getItem('token')
    fetch('/api/plans', { headers: { Authorization: 'Bearer ' + token } })
      .then(r => r.json()).then(setPlans)
  }
  useEffect(() => { load() }, [])

  const create = async (e) => {
    e.preventDefault()
    const token = localStorage.getItem('token')
    await fetch('/api/plans', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
      body: JSON.stringify(form)
    })
    setShowForm(false); load()
  }

  const remove = async (id) => {
    if (!confirm('Delete plan?')) return
    const token = localStorage.getItem('token')
    await fetch('/api/plans/' + id, { method: 'DELETE', headers: { Authorization: 'Bearer ' + token } })
    load()
  }

  const fmt = (bytes) => {
    if (bytes === 0) return 'Unlimited'
    const gb = bytes / 1e9
    return gb >= 1000 ? Math.round(gb/1000) + ' TB' : Math.round(gb) + ' GB'
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Plans</h2>
        <button onClick={() => setShowForm(!showForm)} className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm">+ Add</button>
      </div>
      {showForm && (
        <form onSubmit={create} className="bg-white dark:bg-neutral-800 p-4 rounded-xl shadow-sm mb-4 border grid grid-cols-2 md:grid-cols-5 gap-3">
          <input value={form.name} onChange={e => setForm({...form, name: e.target.value})} placeholder="Name" className="border rounded-lg px-3 py-2 dark:bg-neutral-700" />
          <input type="number" value={form.duration_days} onChange={e => setForm({...form, duration_days: +e.target.value})} placeholder="Days" className="border rounded-lg px-3 py-2 dark:bg-neutral-700" />
          <input type="number" value={form.price} onChange={e => setForm({...form, price: +e.target.value})} placeholder="Price" className="border rounded-lg px-3 py-2 dark:bg-neutral-700" />
          <input type="number" value={form.traffic_limit} onChange={e => setForm({...form, traffic_limit: +e.target.value})} placeholder="Traffic limit" className="border rounded-lg px-3 py-2 dark:bg-neutral-700" />
          <button type="submit" className="bg-green-600 text-white px-4 py-2 rounded-lg">Create</button>
        </form>
      )}
      <div className="bg-white dark:bg-neutral-800 rounded-xl shadow-sm border overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-neutral-700">
            <tr>
              <th className="text-left px-4 py-3">Name</th>
              <th className="text-left px-4 py-3">Duration</th>
              <th className="text-left px-4 py-3">Traffic</th>
              <th className="text-left px-4 py-3">Devices</th>
              <th className="text-left px-4 py-3">Price</th>
              <th className="text-right px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {plans.map(p => (
              <tr key={p.id} className="border-t dark:border-neutral-700">
                <td className="px-4 py-3 font-medium">{p.name}</td>
                <td className="px-4 py-3">{p.duration_days}d</td>
                <td className="px-4 py-3">{fmt(p.traffic_limit)}</td>
                <td className="px-4 py-3">{p.device_limit || 'Unlimited'}</td>
                <td className="px-4 py-3">{p.price} RUB</td>
                <td className="px-4 py-3 text-right"><button onClick={() => remove(p.id)} className="text-red-500 text-xs">Delete</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
