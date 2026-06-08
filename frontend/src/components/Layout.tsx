import { Outlet, NavLink } from "react-router-dom"

const nav = [
  { to: "/", label: "Dashboard", icon: "[S]" },
  { to: "/users", label: "Users", icon: "[U]" },
  { to: "/inbounds", label: "Inbounds", icon: "[I]" },
  { to: "/plans", label: "Plans", icon: "[P]" },
  { to: "/profiles", label: "Clients", icon: "[C]" },
]

export function Layout({ onLogout }: { onLogout: () => void }) {
  return (
    <div className="min-h-screen flex">
      <aside className="w-64 bg-white dark:bg-neutral-900 border-r border-gray-200 dark:border-neutral-700 p-4 flex flex-col">
        <h1 className="text-xl font-bold mb-6 text-blue-600">HsVq Panel</h1>
        <nav className="flex flex-col gap-1 flex-1">
          {nav.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.to === "/"}
              className={({ isActive }) =>
                "flex items-center gap-2 px-3 py-2 rounded-lg transition-colors " +
                (isActive
                  ? "bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300"
                  : "text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-neutral-800")
              }>
              <span>{item.icon}</span>
              {item.label}
            </NavLink>
          ))}
        </nav>
        <button onClick={onLogout} className="mt-4 text-sm text-red-500 hover:text-red-600 text-left px-3 py-2">
          Sign Out
        </button>
      </aside>
      <main className="flex-1 p-6 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
