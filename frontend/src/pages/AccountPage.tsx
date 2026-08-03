import { Link, Navigate } from 'react-router-dom'
import { useEffect, useState, type FormEvent } from 'react'
import { useAuth } from '../context/AuthContext'
import { changeEmail, changePassword } from '../lib/keycloak'
import {
  fetchAccount,
  updateAccount,
  type AccountProfile,
} from '../lib/keycloakAccount'
import {
  authedBlobURL,
  fetchMe,
  getPushSettings,
  getVapidPublicKey,
  setPushSettings,
  subscribePush,
  updateAvatar,
  updateMe,
} from '../lib/api'
import { PrimaryButton, DangerButton, GhostButton } from '../ui/Button'
import { StatusMessage } from '../ui/Card'
import { GlassInput } from '../ui/Input'
import './account.css'

function urlBase64ToUint8Array(base64String: string) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const raw = atob(base64)
  const out = new Uint8Array(raw.length)
  for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i)
  return out
}

export function AccountPage() {
  const auth = useAuth()

  const [profile, setProfile] = useState<AccountProfile | null>(null)
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [nickname, setNickname] = useState('')
  const [avatarUrl, setAvatarUrl] = useState<string | null>(null)
  const [avatarBusy, setAvatarBusy] = useState(false)
  const [pushEnabled, setPushEnabled] = useState(false)

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!auth.enabled || !auth.ready || !auth.authenticated) {
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    void (async () => {
      try {
        const [acc, me, push] = await Promise.all([
          fetchAccount().catch(() => null),
          fetchMe().catch(() => null),
          getPushSettings().catch(() => null),
        ])
        if (cancelled) return
        if (acc) {
          setProfile(acc)
          setFirstName(acc.firstName || '')
          setLastName(acc.lastName || '')
        }
        if (me) {
          setNickname(me.nickname || auth.username || '')
          setAvatarUrl(me.avatarUrl || null)
        } else if (auth.username) {
          setNickname(auth.username)
        }
        if (push) setPushEnabled(Boolean(push.pushEnabled))
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
  }, [auth.enabled, auth.ready, auth.authenticated, auth.username])

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
        email: profile?.email,
      }
      await updateAccount(next)
      const loginNick = (auth.username || nickname).trim()
      if (loginNick) {
        await updateMe(loginNick, `${firstName} ${lastName}`.trim())
      }
      await auth.refreshProfile()
      setNickname(loginNick)
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

  return (
    <div className="accountPage fade-in-up">
      <div className="accountPage__top">
        <Link to="/" className="voco-nav-btn">
          ← На главную
        </Link>
      </div>

      <div className="accountPage__main">
        <div className="accountPage__stack">
          <h1 className="accountPage__title" style={{ fontSize: 28 }}>
            Аккаунт
          </h1>

          {loading && <div className="authBoot" style={{ minHeight: 120 }}>Загрузка…</div>}

          {!loading && (
            <>
              <form className="accountPage__card" onSubmit={onSaveProfile}>
                <h2 className="accountPage__title">Профиль</h2>
                <div className="accountPage__avatarRow">
                  {avatarUrl ? (
                    <img
                      className="accountPage__avatar"
                      src={authedBlobURL(avatarUrl, auth.token)}
                      alt=""
                    />
                  ) : (
                    <div className="accountPage__avatar accountPage__avatar--empty">
                      {(auth.username || '?').slice(0, 2).toUpperCase()}
                    </div>
                  )}
                  <div className="accountPage__avatarActions">
                    <label className="accountPage__avatarBtn">
                      <input
                        type="file"
                        accept="image/*"
                        hidden
                        disabled={avatarBusy}
                        onChange={(e) => {
                          const file = e.target.files?.[0]
                          e.target.value = ''
                          if (!file) return
                          setAvatarBusy(true)
                          setError(null)
                          void updateAvatar(file)
                            .then((u) => {
                              setAvatarUrl(u.avatarUrl || null)
                              setProfileOk('Аватар обновлён.')
                            })
                            .catch((err) =>
                              setError(err instanceof Error ? err.message : 'Не удалось загрузить аватар'),
                            )
                            .finally(() => setAvatarBusy(false))
                        }}
                      />
                      {avatarBusy ? 'Загрузка…' : 'Сменить фото'}
                    </label>
                    <p className="accountPage__hint">JPG/PNG, до 10 МБ</p>
                  </div>
                </div>
                <div className="accountPage__field">
                  <span className="accountPage__label">Логин</span>
                  <div className="accountPage__readonly">{profile?.username ?? auth.username ?? '—'}</div>
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

                <GlassInput
                  label="Никнейм (логин)"
                  value={auth.username || nickname}
                  readOnly
                  autoComplete="username"
                />
                <p className="accountPage__hint">Так вас будут видеть другие юзеры. Совпадает с логином.</p>

                {profileOk && <div className="accountPage__ok">{profileOk}</div>}

                <PrimaryButton type="submit" loading={profileBusy}>
                  {profileBusy ? 'Сохраняю…' : 'Сохранить профиль'}
                </PrimaryButton>
              </form>

              <div className="accountPage__card">
                <h2 className="accountPage__title">Уведомления</h2>
                <div className="accountPage__actions" style={{ justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ color: 'var(--voco-text)', fontSize: 14 }}>
                    Browser push: {pushEnabled ? 'вкл' : 'выкл'}
                  </span>
                  <GhostButton
                    type="button"
                    onClick={() => {
                      void (async () => {
                        try {
                          const next = !pushEnabled
                          if (next) {
                            const perm = await Notification.requestPermission()
                            if (perm !== 'granted') throw new Error('Разрешение не выдано')
                            const reg = await navigator.serviceWorker.register('/sw.js')
                            const { publicKey } = await getVapidPublicKey()
                            if (!publicKey) throw new Error('VAPID public key не настроен на сервере')
                            const sub = await reg.pushManager.subscribe({
                              userVisibleOnly: true,
                              applicationServerKey: urlBase64ToUint8Array(publicKey),
                            })
                            await subscribePush(sub.toJSON())
                            await setPushSettings(true)
                            setPushEnabled(true)
                          } else {
                            await setPushSettings(false)
                            setPushEnabled(false)
                          }
                        } catch (e) {
                          setError(e instanceof Error ? e.message : 'Не удалось изменить push')
                        }
                      })()
                    }}
                  >
                    {pushEnabled ? 'Выключить' : 'Включить'}
                  </GhostButton>
                </div>
                <p className="accountPage__hint">
                  По умолчанию выключено. In-app toast работает при открытой вкладке через WebSocket.
                </p>
              </div>

              <div className="accountPage__card">
                <h2 className="accountPage__title">Email</h2>
                <div className="accountPage__field">
                  <span className="accountPage__label">Текущий email</span>
                  <div className="accountPage__readonly">
                    {profile?.email ?? '—'}
                    {profile?.emailVerified === false ? ' (не подтверждён)' : ''}
                  </div>
                </div>
                {emailOk && <div className="accountPage__ok">{emailOk}</div>}
                <div className="accountPage__actions accountPage__actions--stack">
                  <PrimaryButton type="button" loading={emailBusy} onClick={() => void onChangeEmail()}>
                    {emailBusy ? 'Перехожу…' : 'Сменить email'}
                  </PrimaryButton>
                </div>
              </div>

              <div className="accountPage__card">
                <h2 className="accountPage__title">Смена пароля</h2>
                {passwordOk && <div className="accountPage__ok">{passwordOk}</div>}
                <div className="accountPage__actions accountPage__actions--stack">
                  <PrimaryButton type="button" loading={passwordBusy} onClick={() => void onChangePassword()}>
                    {passwordBusy ? 'Перехожу…' : 'Сменить пароль'}
                  </PrimaryButton>
                </div>
              </div>

              <div className="accountPage__card">
                <h2 className="accountPage__title">Сессия</h2>
                <DangerButton type="button" fullWidth onClick={() => void auth.logout()}>
                  Выйти
                </DangerButton>
              </div>
            </>
          )}

          {error && <StatusMessage type="error">{error}</StatusMessage>}
        </div>
      </div>
    </div>
  )
}
