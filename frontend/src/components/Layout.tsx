import { Outlet, NavLink } from 'react-router-dom'

const nav = [
  { to: '/', label: 'Dashboard', icon: '◆' },
  { to: '/users', label: 'Users', icon: '●' },
  { to: '/inbounds', label: 'Inbounds', icon: '▣' },
  { to: '/plans', label: 'Plans', icon: '▲' },
  { to: '/profiles', label: 'Clients', icon: '★' },
]

export function Layout({ onLogout }: { onLogout: () => void }) {
  return (
    <div className="h-screen flex bg-gray-100">
      <aside className="w-56 bg-white shadow-sm flex flex-col">
        <div className="p-5 border-b">
          <h1 className="text-lg font-bold text-blue-600">HsVq Panel</h1>
        </div>
        <nav className="flex-1 p-3 space-y-1">
          {nav.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.to === '/'}
              className={({ isActive }) =>
                'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all duration-150 active:scale-[0.97] ' +
                (isActive
                  ? 'bg-blue-50 text-blue-700 font-medium shadow-sm'
                  : 'text-gray-600 hover:bg-gray-100')
              }>
              <span className="text-base w-5 text-center">{item.icon}</span>
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="p-3 border-t">
          <button onClick={onLogout}
            className="w-full text-left px-3 py-2 text-sm text-red-500 hover:bg-red-50 rounded-lg transition-all active:scale-[0.97]">
            Sign Out
          </button>
        </div>
      </aside>
      <main className="flex-1 overflow-auto p-6">
        <Outlet />
      </main>
    </div>
  )
}
