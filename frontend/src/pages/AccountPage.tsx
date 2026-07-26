import { Link, Navigate } from 'react-router-dom'
import { useEffect, useState, type CSSProperties, type FormEvent } from 'react'
import { useAuth } from '../context/AuthContext'
import { changeEmail, changePassword } from '../lib/keycloak'
import {
  fetchAccount,
  updateAccount,
  type AccountProfile,
} from '../lib/keycloakAccount'
import { PrimaryButton, NavButton } from '../ui/Button'
import { StatusMessage } from '../ui/Card'
import { GlassInput } from '../ui/Input'

export function AccountPage() {
  const auth = useAuth()

  const [profile, setProfile] = useState<AccountProfile | null>(null)
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')

  const [loading, setLoading] = useState(true)
  const [profileBusy, setProfileBusy] = useState(false)
  const [passwordBusy, setPasswordBusy] = useState(false)
  const [emailBusy, setEmailBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [profileOk, setProfileOk] = useState<string | null>(null)
  const [passwordOk, setPasswordOk] = useState<string | null>(null)
  const [emailOk, setEmailOk] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const hashParams = new URLSearchParams(window.location.hash.replace(/^#/, ''))
    const status = params.get('kc_action_status') ?? hashParams.get('kc_action_status')
    const action = params.get('kc_action') ?? hashParams.get('kc_action')
    if (status === 'success') {
      if (action === 'UPDATE_EMAIL') {
        setEmailOk('Email обновлён. Проверь новое письмо, если Keycloak запросил подтверждение.')
        void auth.refreshProfile()
        void fetchAccount()
          .then((data) => {
            setProfile(data)
          })
          .catch(() => {
            /* ignore */
          })
      } else {
        setPasswordOk('Пароль успешно изменён.')
      }
    } else if (status === 'cancelled') {
      setError(action === 'UPDATE_EMAIL' ? 'Смена email отменена.' : 'Смена пароля отменена.')
    }
    // auth.refreshProfile is stable enough; avoid re-running on every auth change
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!auth.enabled || !auth.ready || !auth.authenticated) {
      setLoading(false)
      return
    }

    let cancelled = false
    void (async () => {
      try {
        const data = await fetchAccount()
        if (cancelled) return
        setProfile(data)
        setFirstName(data.firstName ?? '')
        setLastName(data.lastName ?? '')
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : 'Не удалось загрузить профиль')
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()

    return () => {
      cancelled = true
    }
  }, [auth.enabled, auth.ready, auth.authenticated])

  if (auth.enabled && auth.ready && !auth.authenticated) {
    return <Navigate to="/" replace />
  }

  if (!auth.enabled) {
    return <Navigate to="/" replace />
  }

  async function onSaveProfile(e: FormEvent) {
    e.preventDefault()
    setProfileBusy(true)
    setError(null)
    setProfileOk(null)
    try {
      const next: AccountProfile = {
        ...profile,
        username: profile?.username,
        firstName: firstName.trim(),
        lastName: lastName.trim(),
        // Email меняется только через UPDATE_EMAIL AIA (письмо на новый адрес).
        email: profile?.email,
      }
      await updateAccount(next)
      await auth.refreshProfile()
      setProfile(next)
      setProfileOk('Профиль успешно сохранён.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить профиль')
    } finally {
      setProfileBusy(false)
    }
  }

  async function onChangePassword() {
    setPasswordBusy(true)
    setError(null)
    setPasswordOk(null)
    try {
      await changePassword(`${window.location.origin}/account`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось открыть смену пароля')
      setPasswordBusy(false)
    }
  }

  async function onChangeEmail() {
    setEmailBusy(true)
    setError(null)
    setEmailOk(null)
    try {
      await changeEmail(`${window.location.origin}/account`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось открыть смену email')
      setEmailBusy(false)
    }
  }

  const sectionStyle: CSSProperties = {
    padding: 28,
    borderRadius: 16,
    background: 'var(--voco-card)',
    backdropFilter: 'blur(20px)',
    border: '1px solid var(--voco-border-soft)',
    display: 'flex',
    flexDirection: 'column',
    gap: 16,
  }

  return (
    <div
      className="fade-in-up"
      style={{
        position: 'relative',
        zIndex: 1,
        minHeight: '100vh',
        padding: '32px 24px 80px',
        maxWidth: 600,
        margin: '0 auto',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 44, gap: 12 }}>
        <Link to="/" className="voco-nav-btn">
          ← На главную
        </Link>
        <NavButton onClick={() => void auth.logout()}>Выйти</NavButton>
      </div>

      <h1
        style={{
          fontFamily: 'Outfit, sans-serif',
          fontWeight: 600,
          fontSize: 36,
          color: 'var(--voco-text)',
          letterSpacing: '-0.01em',
          marginBottom: 40,
        }}
      >
        Аккаунт
      </h1>

      {loading && <div className="authBoot" style={{ minHeight: 120 }}>Загрузка…</div>}

      {!loading && (
        <>
          <section style={{ marginBottom: 40 }}>
            <SectionLabel>Профиль</SectionLabel>
            <form style={sectionStyle} onSubmit={onSaveProfile}>
              <div
                style={{
                  padding: '10px 14px',
                  borderRadius: 10,
                  background: 'var(--voco-chip-bg)',
                  border: '1px solid var(--voco-border-soft)',
                }}
              >
                <p
                  style={{
                    fontFamily: 'Outfit, sans-serif',
                    fontSize: 11,
                    color: 'var(--voco-text-faint)',
                    letterSpacing: '0.08em',
                    textTransform: 'uppercase',
                    margin: '0 0 4px',
                  }}
                >
                  Логин
                </p>
                <p style={{ fontFamily: 'Outfit, sans-serif', fontSize: 14, color: 'var(--voco-text-muted)', margin: 0 }}>
                  {profile?.username ?? auth.username ?? '—'}
                </p>
              </div>

              <div style={{ display: 'flex', gap: 12 }}>
                <GlassInput
                  label="Имя"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  className="flex-1"
                  autoComplete="given-name"
                />
                <GlassInput
                  label="Фамилия"
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                  className="flex-1"
                  autoComplete="family-name"
                />
              </div>

              {profileOk && <StatusMessage type="success">{profileOk}</StatusMessage>}

              <PrimaryButton type="submit" loading={profileBusy}>
                {profileBusy ? 'Сохраняю…' : 'Сохранить профиль'}
              </PrimaryButton>
            </form>
          </section>

          <section style={{ marginBottom: 40 }}>
            <SectionLabel>Email</SectionLabel>
            <div style={sectionStyle}>
              <div
                style={{
                  padding: '10px 14px',
                  borderRadius: 10,
                  background: 'var(--voco-chip-bg)',
                  border: '1px solid var(--voco-border-soft)',
                }}
              >
                <p
                  style={{
                    fontFamily: 'Outfit, sans-serif',
                    fontSize: 11,
                    color: 'var(--voco-text-faint)',
                    letterSpacing: '0.08em',
                    textTransform: 'uppercase',
                    margin: '0 0 4px',
                  }}
                >
                  Текущий email
                </p>
                <p style={{ fontFamily: 'Outfit, sans-serif', fontSize: 14, color: 'var(--voco-text-muted)', margin: 0 }}>
                  {profile?.email ?? '—'}
                  {profile?.emailVerified === false ? ' (не подтверждён)' : ''}
                </p>
              </div>
              {emailOk && <StatusMessage type="success">{emailOk}</StatusMessage>}
              <PrimaryButton type="button" loading={emailBusy} onClick={() => void onChangeEmail()}>
                {emailBusy ? 'Перехожу…' : 'Сменить email'}
              </PrimaryButton>
            </div>
          </section>

          <section>
            <SectionLabel>Смена пароля</SectionLabel>
            <div style={sectionStyle}>
              {passwordOk && <StatusMessage type="success">{passwordOk}</StatusMessage>}
              <PrimaryButton type="button" loading={passwordBusy} onClick={() => void onChangePassword()}>
                {passwordBusy ? 'Перехожу…' : 'Сменить пароль'}
              </PrimaryButton>
            </div>
          </section>
        </>
      )}

      {error && (
        <div style={{ marginTop: 24 }}>
          <StatusMessage type="error">{error}</StatusMessage>
        </div>
      )}
    </div>
  )
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <p
      style={{
        fontFamily: 'Outfit, sans-serif',
        fontSize: 11,
        fontWeight: 500,
        letterSpacing: '0.12em',
        color: 'var(--voco-text-faint)',
        textTransform: 'uppercase',
        marginBottom: 12,
      }}
    >
      {children}
    </p>
  )
}
