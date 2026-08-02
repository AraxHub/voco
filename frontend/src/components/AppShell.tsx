import { NavLink, Outlet, Navigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import './AppShell.css'

const links = [
  { to: '/', label: 'Главная', end: true },
  { to: '/chats', label: 'Чаты' },
  { to: '/calendar', label: 'Календарь' },
]

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const auth = useAuth()
  if (!auth.ready) return <div className="authBoot">Загрузка…</div>
  if (auth.enabled && !auth.authenticated) {
    return <Navigate to="/" replace />
  }
  return <>{children}</>
}

export function AppShell() {
  const auth = useAuth()
  const initials = (auth.displayName || auth.username || 'U').slice(0, 2).toUpperCase()

  return (
    <div className="appShell">
      <aside className="appShell__rail" aria-label="Навигация">
        <div className="appShell__brand">VOCO</div>
        <nav className="appShell__nav">
          {links.map((l) => (
            <NavLink
              key={l.to}
              to={l.to}
              end={l.end}
              className={({ isActive }) => `appShell__link${isActive ? ' is-active' : ''}`}
            >
              {l.label}
            </NavLink>
          ))}
        </nav>
        <NavLink to="/account" className="appShell__avatar" title="Аккаунт">
          {initials}
        </NavLink>
      </aside>
      <main className="appShell__main">
        <Outlet />
      </main>
      <nav className="appShell__bottom" aria-label="Мобильная навигация">
        {links.map((l) => (
          <NavLink
            key={l.to}
            to={l.to}
            end={l.end}
            className={({ isActive }) => `appShell__bottomLink${isActive ? ' is-active' : ''}`}
          >
            {l.label}
          </NavLink>
        ))}
        <NavLink to="/account" className="appShell__bottomLink">
          Аккаунт
        </NavLink>
      </nav>
    </div>
  )
}
