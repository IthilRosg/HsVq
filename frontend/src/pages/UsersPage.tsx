import { useEffect, useState } from 'react'
import { apiGetUsers, apiCreateUser, apiDeleteUser } from '../lib/api'

export function UsersPage() {
  const [users, setUsers] = useState([])
  const [showForm, setShowForm] = useState(false)
  const [email, setEmail] = useState('')
  const [pass, setPass] = useState('')
  const load = () => apiGetUsers().then(setUsers)
  useEffect(() => { load() }, [])
  const create = async (e) => {
    e.preventDefault()
    await apiCreateUser({ email, password_hash: pass })
    setShowForm(false); setEmail(''); setPass(''); load()
  }
  const remove = async (id) => {
    if (!confirm('Delete?')) return
    await apiDeleteUser(id); load()
  }
  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Users</h2>
        <button onClick={() => setShowForm(!showForm)} className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm">+ Add</button>
      </div>
      {showForm && (
        <form onSubmit={create} className="bg-white dark:bg-neutral-800 p-4 rounded-xl shadow-sm mb-4 border flex gap-3 items-end">
          <input value={email} onChange={e => setEmail(e.target.value)} placeholder="Email"
            className="border rounded-lg px-3 py-2 flex-1 dark:bg-neutral-700" />
          <input value={pass} onChange={e => setPass(e.target.value)} placeholder="Password" type="password"
            className="border rounded-lg px-3 py-2 flex-1 dark:bg-neutral-700" />
          <button type="submit" className="bg-green-600 text-white px-4 py-2 rounded-lg">Create</button>
        </form>
      )}
      <div className="bg-white dark:bg-neutral-800 rounded-xl shadow-sm border overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-neutral-700">
            <tr>
              <th className="text-left px-4 py-3">ID</th>
              <th className="text-left px-4 py-3">Email</th>
              <th className="text-left px-4 py-3">Role</th>
              <th className="text-left px-4 py-3">Status</th>
              <th className="text-right px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {users.map(u => (
              <tr key={u.id} className="border-t dark:border-neutral-700">
                <td className="px-4 py-3 font-mono text-xs">{u.id}</td>
                <td className="px-4 py-3">{u.email}</td>
                <td className="px-4 py-3"><span className="px-2 py-0.5 rounded text-xs bg-purple-100 text-purple-700">{u.role}</span></td>
                <td className="px-4 py-3"><span className="px-2 py-0.5 rounded text-xs bg-green-100 text-green-700">{u.status}</span></td>
                <td className="px-4 py-3 text-right"><button onClick={() => remove(u.id)} className="text-red-500 text-xs">Delete</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
