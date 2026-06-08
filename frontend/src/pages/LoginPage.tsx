import { useState } from 'react'
import { apiLogin } from '../lib/api'

export function LoginPage({ onLogin }: { onLogin: (token: string) => void }) {
  const [email, setEmail] = useState('admin')
  const [password, setPassword] = useState('admin')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const data = await apiLogin(email, password)
      onLogin(data.token)
    } catch {
      setError('Invalid credentials')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-neutral-950">
      <div className="bg-white dark:bg-neutral-900 p-8 rounded-xl shadow-md w-full max-w-sm">
        <h1 className="text-2xl font-bold text-center mb-6 text-blue-600">HsVq Panel</h1>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <input className="border border-gray-300 dark:border-neutral-600 rounded-lg px-3 py-2 bg-white dark:bg-neutral-800"
            placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} />
          <input type="password" className="border border-gray-300 dark:border-neutral-600 rounded-lg px-3 py-2 bg-white dark:bg-neutral-800"
            placeholder="Password" value={password} onChange={e => setPassword(e.target.value)} />
          {error && <p className="text-red-500 text-sm">{error}</p>}
          <button type="submit" disabled={loading}
            className="bg-blue-600 hover:bg-blue-700 text-white rounded-lg px-4 py-2 font-medium disabled:opacity-50">
            {loading ? 'Loading...' : 'Sign In'}
          </button>
        </form>
      </div>
    </div>
  )
}
