import { NavLink, Outlet, Navigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { keycloak } from '../lib/keycloak'
import './AppShell.css'

const links = [
  {
    to: '/',
    label: 'Главная',
    end: true,
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M12.4 3.2a1.2 1.2 0 0 0-1.5.05l-7.2 6.5A1.2 1.2 0 0 0 4.2 12H5.5v6.3c0 .9.7 1.6 1.6 1.6H10v-4.6c0-.7.5-1.2 1.2-1.2h1.6c.7 0 1.2.5 1.2 1.2v4.6h2.9c.9 0 1.6-.7 1.6-1.6V12h1.3a1.2 1.2 0 0 0 .5-2.25l-7.4-6.55Z"
        />
      </svg>
    ),
  },
  {
    to: '/chats',
    label: 'Чаты',
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M12 3.2c-4.8 0-8.8 3.3-8.8 7.4 0 2.4 1.3 4.5 3.4 5.9-.1.8-.5 2.1-1.6 3.3-.2.2 0 .6.3.5 1.8-.5 3.2-1.4 4.1-2.1.8.2 1.7.3 2.6.3 4.8 0 8.8-3.3 8.8-7.4S16.8 3.2 12 3.2Zm-3.2 6.6a1.15 1.15 0 1 1 0 2.3 1.15 1.15 0 0 1 0-2.3Zm3.2 0a1.15 1.15 0 1 1 0 2.3 1.15 1.15 0 0 1 0-2.3Zm3.2 0a1.15 1.15 0 1 1 0 2.3 1.15 1.15 0 0 1 0-2.3Z"
        />
      </svg>
    ),
  },
  {
    to: '/calendar',
    label: 'Календарь',
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M8.2 3.4a1 1 0 0 0-1 1V5H6.4A2.4 2.4 0 0 0 4 7.4v11.2A2.4 2.4 0 0 0 6.4 21h10.4a1.2 1.2 0 0 0 .9-.4l1.8-2a1.2 1.2 0 0 0 .3-.8V7.4A2.4 2.4 0 0 0 17.6 5h-.8V4.4a1 1 0 1 0-2 0V5H10.2V4.4a1 1 0 0 0-1-1h-1Zm-1.8 6h11.2v8.8H7.2v-1.6a1.2 1.2 0 0 0-.4-.9l-1.2-1.1V9.4Zm2.4 2.2a1 1 0 1 0 0 2h1.8a1 1 0 1 0 0-2H9.6Zm4.2 0a1 1 0 1 0 0 2h1.6a1 1 0 1 0 0-2h-1.6Z"
        />
      </svg>
    ),
  },
]

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const auth = useAuth()
  if (!auth.ready) return <div className="authBoot">Загрузка…</div>
  // Prefer live Keycloak instance — avoids a one-frame false negative after SSO restore.
  const ok = auth.authenticated || Boolean(keycloak?.authenticated)
  if (auth.enabled && !ok) {
    return <Navigate to="/" replace />
  }
  return <>{children}</>
}

export function AppShell() {
  const auth = useAuth()
  const initials = (auth.displayName || auth.username || 'U').slice(0, 2).toUpperCase()
  const showNav = auth.ready && (!auth.enabled || auth.authenticated)

  if (!showNav) {
    return (
      <div className="appShell appShell--guest">
        <main className="appShell__main">
          <Outlet />
        </main>
      </div>
    )
  }

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
              <span className="appShell__icon">{l.icon}</span>
              <span className="appShell__label">{l.label}</span>
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
            <span className="appShell__icon">{l.icon}</span>
            <span className="appShell__label">{l.label}</span>
          </NavLink>
        ))}
        <NavLink to="/account" className="appShell__bottomLink">
          <span className="appShell__icon appShell__icon--text">{initials}</span>
          <span className="appShell__label">Аккаунт</span>
        </NavLink>
      </nav>
    </div>
  )
}
