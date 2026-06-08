import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useState } from 'react'
import { LoginPage } from './pages/LoginPage'
import { DashboardPage } from './pages/DashboardPage'
import { UsersPage } from './pages/UsersPage'
import { InboundsPage } from './pages/InboundsPage'
import { PlansPage } from './pages/PlansPage'
import { ProfilesPage } from './pages/ProfilesPage'
import { Layout } from './components/Layout'

function App() {
  const [token, setToken] = useState(localStorage.getItem('token'))

  const onLogin = (t: string) => {
    localStorage.setItem('token', t)
    setToken(t)
  }

  const onLogout = () => {
    localStorage.removeItem('token')
    setToken(null)
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={token ? <Navigate to="/" /> : <LoginPage onLogin={onLogin} />} />
        <Route path="/" element={token ? <Layout onLogout={onLogout} /> : <Navigate to="/login" />}>
          <Route index element={<DashboardPage />} />
          <Route path="users" element={<UsersPage />} />
          <Route path="inbounds" element={<InboundsPage />} />
          <Route path="plans" element={<PlansPage />} />
          <Route path="profiles" element={<ProfilesPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
