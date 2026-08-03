import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  acceptRequest,
  addGroupMember,
  authedBlobURL,
  blockRequest,
  callFromChat,
  createGroup,
  fetchMe,
  getConversationRequest,
  listConversations,
  listMembers,
  listMessages,
  markRead,
  openDirect,
  renameGroup,
  searchUsers,
  sendMessage,
  sendTyping,
  type Conversation,
  type ConversationMember,
  type Message,
  type VocoUser,
  wsURL,
} from '../lib/api'
import { useAuth } from '../context/AuthContext'
import { PrimaryButton, GhostButton } from '../ui/Button'
import { GlassInput } from '../ui/Input'
import './chats.css'

function cid(c: Conversation) {
  return (c.id || c.ID || '').toString()
}
function mid(m: Message) {
  return (m.id || m.ID || '').toString()
}
function titleOf(c: Conversation) {
  return c.title || c.Title || 'Чат'
}
function typeOf(c: Conversation) {
  return (c.type || c.Type || '').toString()
}
function senderLabel(m: Message) {
  return m.senderName || 'Участник'
}
function senderId(m: Message) {
  return (m.senderId || m.SenderID || '').toString()
}

function initials(title: string) {
  const t = title.trim()
  if (!t) return '?'
  const parts = t.split(/\s+/).filter(Boolean)
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase()
  return t.slice(0, 2).toUpperCase()
}

function formatListTime(iso?: string) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const now = new Date()
  const sameDay = d.toDateString() === now.toDateString()
  if (sameDay) return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  return d.toLocaleDateString([], { day: 'numeric', month: 'short' })
}

export function ChatsPage() {
  const auth = useAuth()
  const nav = useNavigate()
  const { conversationId } = useParams()
  const [list, setList] = useState<Conversation[]>([])
  const [messages, setMessages] = useState<Message[]>([])
  const [members, setMembers] = useState<ConversationMember[]>([])
  const [me, setMe] = useState<VocoUser | null>(null)
  const [q, setQ] = useState('')
  const [found, setFound] = useState<VocoUser[]>([])
  const [text, setText] = useState('')
  const [pendingFiles, setPendingFiles] = useState<File[]>([])
  const [groupTitle, setGroupTitle] = useState('')
  const [groupPick, setGroupPick] = useState<VocoUser[]>([])
  const [editTitle, setEditTitle] = useState('')
  const [showGroupEdit, setShowGroupEdit] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)
  const [incomingPending, setIncomingPending] = useState(false)
  const [typingLabel, setTypingLabel] = useState<string | null>(null)
  const [lightbox, setLightbox] = useState<string | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)
  const typingTimer = useRef<number | null>(null)
  const lastTypingSent = useRef(0)

  const active = useMemo(
    () => list.find((c) => cid(c) === conversationId) || null,
    [list, conversationId],
  )
  const activeIsGroup = active ? typeOf(active) === 'group' : false
  const myId = me?.id || ''
  const myNames = useMemo(() => {
    const names = [auth.username, auth.displayName, me?.nickname, me?.displayName]
      .filter(Boolean)
      .map((s) => s!.toLowerCase())
    return new Set(names)
  }, [auth.username, auth.displayName, me])

  function isOwnMessage(m: Message) {
    const sid = senderId(m)
    if (myId && sid && sid === myId) return true
    const name = (m.senderName || '').toLowerCase()
    return Boolean(name && myNames.has(name))
  }

  function mediaURL(url: string) {
    return authedBlobURL(url, auth.token)
  }

  async function reloadList() {
    const data = await listConversations()
    setList(Array.isArray(data) ? data : [])
  }

  async function reloadRequest(id: string) {
    try {
      const state = await getConversationRequest(id)
      setIncomingPending(Boolean(state.incomingPending))
    } catch {
      setIncomingPending(false)
    }
  }

  async function reloadMembers(id: string) {
    try {
      const data = await listMembers(id)
      setMembers(Array.isArray(data) ? data : [])
    } catch {
      setMembers([])
    }
  }

  useEffect(() => {
    void fetchMe()
      .then(setMe)
      .catch(() => undefined)
    void reloadList().catch((e) => setError(String(e)))
  }, [])

  useEffect(() => {
    const refresh = () => {
      void reloadList().catch(() => undefined)
    }
    window.addEventListener('focus', refresh)
    const onVis = () => {
      if (document.visibilityState === 'visible') refresh()
    }
    document.addEventListener('visibilitychange', onVis)
    return () => {
      window.removeEventListener('focus', refresh)
      document.removeEventListener('visibilitychange', onVis)
    }
  }, [])

  useEffect(() => {
    if (!conversationId) {
      setMessages([])
      setMembers([])
      setIncomingPending(false)
      setShowGroupEdit(false)
      setTypingLabel(null)
      setPendingFiles([])
      return
    }
    void listMessages(conversationId)
      .then((m) => {
        const arr = Array.isArray(m) ? m : []
        setMessages(arr)
        const newest = arr[0]
        if (newest) {
          void markRead(conversationId, mid(newest)).then(() => reloadList()).catch(() => undefined)
        }
      })
      .catch((e) => setError(String(e)))
    void reloadRequest(conversationId)
    void reloadMembers(conversationId)
  }, [conversationId])

  useEffect(() => {
    if (active && activeIsGroup) {
      setEditTitle(titleOf(active))
    }
  }, [active, activeIsGroup])

  useEffect(() => {
    if (!auth.token) return
    let ws: WebSocket
    try {
      ws = new WebSocket(wsURL(auth.token))
    } catch {
      return
    }
    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(String(ev.data)) as {
          event?: string
          payload?: {
            Body?: string
            title?: string
            Title?: string
            conversationId?: string
            ConversationID?: string
            userId?: string
            UserID?: string
            displayName?: string
            nickname?: string
          }
        }
        if (msg.event === 'message.created') {
          void reloadList()
          if (conversationId) {
            void listMessages(conversationId).then((m) => {
              setMessages(m)
              const newest = m[0]
              if (newest) void markRead(conversationId, mid(newest)).then(() => reloadList())
            })
            void reloadRequest(conversationId)
          }
          setToast('Новое сообщение')
        }
        if (msg.event === 'conversation.updated' || msg.event === 'conversation.created') {
          void reloadList()
          if (conversationId) {
            void reloadMembers(conversationId)
            void reloadRequest(conversationId)
          }
          const title = msg.payload?.title || msg.payload?.Title
          if (title) setToast(String(title))
        }
        if (msg.event === 'typing') {
          const cidEvt = (msg.payload?.conversationId || msg.payload?.ConversationID || '').toString()
          if (cidEvt && conversationId && cidEvt === conversationId) {
            const uid = (msg.payload?.userId || msg.payload?.UserID || '').toString()
            if (uid && myId && uid === myId) return
            const name =
              msg.payload?.displayName ||
              msg.payload?.nickname ||
              members.find((m) => m.userId === uid)?.name ||
              'Кто-то'
            setTypingLabel(`${name} печатает…`)
            if (typingTimer.current) window.clearTimeout(typingTimer.current)
            typingTimer.current = window.setTimeout(() => setTypingLabel(null), 2500)
          }
        }
        if (msg.event === 'notification') {
          setToast(msg.payload?.title || 'Уведомление')
        }
      } catch {
        /* ignore */
      }
    }
    return () => {
      ws.close()
      if (typingTimer.current) window.clearTimeout(typingTimer.current)
    }
  }, [auth.token, conversationId, myId, members])

  useEffect(() => {
    if (!toast) return
    const t = window.setTimeout(() => setToast(null), 3200)
    return () => window.clearTimeout(t)
  }, [toast])

  async function onSearch(value: string) {
    setQ(value)
    if (value.trim().length < 1) {
      setFound([])
      return
    }
    try {
      setFound(await searchUsers(value.trim()))
    } catch (e) {
      setError(String(e))
    }
  }

  function onComposerChange(value: string) {
    setText(value)
    if (!conversationId) return
    const now = Date.now()
    if (now - lastTypingSent.current < 1200) return
    lastTypingSent.current = now
    void sendTyping(conversationId).catch(() => undefined)
  }

  function previewBody(c: Conversation) {
    return c.lastMessage?.body || 'Нет сообщений'
  }

  return (
    <div className="chats">
      <aside className="chats__sidebar">
        <GlassInput value={q} onChange={(e) => void onSearch(e.target.value)} placeholder="Поиск по нику" />
        {found.length > 0 && (
          <div className="chats__searchHits">
            {found.map((u) => (
              <div key={u.id} className="chats__hitRow">
                <button
                  type="button"
                  className="chats__hit"
                  onClick={() => {
                    void openDirect(u.id).then((r) => {
                      const id = cid(r.conversation)
                      void reloadList()
                      nav(`/chats/${id}`)
                      setFound([])
                      setQ('')
                    })
                  }}
                >
                  {u.avatarUrl ? (
                    <img className="chats__hitAvatar" src={mediaURL(u.avatarUrl)} alt="" />
                  ) : (
                    <span className="chats__hitAvatar chats__hitAvatar--empty">{initials(u.displayName || u.nickname || '?')}</span>
                  )}
                  <span>@{u.nickname || u.email}</span>
                </button>
                <GhostButton
                  type="button"
                  onClick={() => {
                    setGroupPick((prev) => (prev.some((x) => x.id === u.id) ? prev : [...prev, u]))
                    setFound([])
                    setQ('')
                  }}
                >
                  В группу
                </GhostButton>
              </div>
            ))}
          </div>
        )}
        <div className="chats__newGroup">
          <GlassInput value={groupTitle} onChange={(e) => setGroupTitle(e.target.value)} placeholder="Название группы" />
          {groupPick.length > 0 && (
            <div className="chats__picked">
              {groupPick.map((u) => (
                <button
                  key={u.id}
                  type="button"
                  className="chats__chip"
                  onClick={() => setGroupPick((prev) => prev.filter((x) => x.id !== u.id))}
                >
                  {u.displayName || u.nickname} ×
                </button>
              ))}
            </div>
          )}
          <GhostButton
            type="button"
            onClick={() => {
              const trimmed = groupTitle.trim()
              if (!trimmed) {
                setError('Укажите название группы')
                return
              }
              setError(null)
              void createGroup(
                trimmed,
                groupPick.map((u) => u.id),
              )
                .then((c) => {
                  void reloadList()
                  nav(`/chats/${cid(c)}`)
                  setGroupTitle('')
                  setGroupPick([])
                })
                .catch((e) => {
                  const msg = (e instanceof Error ? e.message : String(e)).replace(/^Error:\s*/i, '')
                  if (/сессия истекла|invalid token|unauthorized|authorization required/i.test(msg)) {
                    setError('Сессия истекла или токен недействителен. Войдите снова.')
                    return
                  }
                  setError(/^validation$/i.test(msg) ? 'Укажите название группы' : msg)
                })
            }}
          >
            + Группа{groupPick.length ? ` (${groupPick.length})` : ''}
          </GhostButton>
        </div>
        <div className="chats__list">
          {list.map((c) => (
            <button
              key={cid(c)}
              type="button"
              className={`chats__item${cid(c) === conversationId ? ' is-active' : ''}`}
              onClick={() => nav(`/chats/${cid(c)}`)}
            >
              <span className="chats__itemAvatarWrap">
                {c.avatarUrl ? (
                  <img className="chats__itemAvatar" src={mediaURL(c.avatarUrl)} alt="" />
                ) : (
                  <span className="chats__itemAvatar chats__itemAvatar--empty">{initials(titleOf(c))}</span>
                )}
              </span>
              <span className="chats__itemMain">
                <span className="chats__itemTop">
                  <span className="chats__itemTitle">{titleOf(c)}</span>
                  <span className="chats__itemTime">{formatListTime(c.lastMessage?.createdAt)}</span>
                </span>
                <span className="chats__itemBottom">
                  <span className="chats__itemPreview">{previewBody(c)}</span>
                  {(c.unreadCount || 0) > 0 && <span className="chats__badge">{c.unreadCount}</span>}
                </span>
              </span>
            </button>
          ))}
        </div>
      </aside>
      <section className="chats__thread">
        {!active ? (
          <div className="chats__empty">Выберите, кому вы хотите написать</div>
        ) : (
          <>
            <header className="chats__head">
              <div className="chats__headTitle">
                {active.avatarUrl ? (
                  <img className="chats__headAvatar" src={mediaURL(active.avatarUrl)} alt="" />
                ) : (
                  <span className="chats__headAvatar chats__headAvatar--empty">{initials(titleOf(active))}</span>
                )}
                <div>
                  <h2>{titleOf(active)}</h2>
                  {typingLabel && <div className="chats__typing">{typingLabel}</div>}
                </div>
              </div>
              <div className="chats__actions">
                {activeIsGroup && (
                  <GhostButton type="button" onClick={() => setShowGroupEdit((v) => !v)}>
                    {showGroupEdit ? 'Скрыть' : 'Участники'}
                  </GhostButton>
                )}
                {incomingPending && messages.length > 0 && (
                  <>
                    <GhostButton
                      onClick={() =>
                        void acceptRequest(cid(active))
                          .then(() => {
                            setIncomingPending(false)
                            setToast('Запрос принят')
                          })
                          .catch((e) => setError(String(e)))
                      }
                    >
                      Принять
                    </GhostButton>
                    <GhostButton
                      onClick={() =>
                        void blockRequest(cid(active))
                          .then(() => {
                            setIncomingPending(false)
                            setToast('Заблокировано')
                          })
                          .catch((e) => setError(String(e)))
                      }
                    >
                      Заблокировать
                    </GhostButton>
                  </>
                )}
                <PrimaryButton
                  onClick={() =>
                    void callFromChat(cid(active))
                      .then((r) => nav(`/room/${r.roomId}?awaiting=1&join=1`))
                      .catch((e) => setError(String(e)))
                  }
                >
                  Позвонить
                </PrimaryButton>
              </div>
            </header>
            {showGroupEdit && activeIsGroup && conversationId && (
              <div className="chats__groupEdit">
                <form
                  className="chats__groupRename"
                  onSubmit={(e) => {
                    e.preventDefault()
                    const t = editTitle.trim()
                    if (!t) return
                    void renameGroup(conversationId, t)
                      .then(() => {
                        setToast('Название обновлено')
                        return reloadList()
                      })
                      .catch((err) => setError(String(err)))
                  }}
                >
                  <GlassInput value={editTitle} onChange={(e) => setEditTitle(e.target.value)} placeholder="Название группы" />
                  <GhostButton type="submit">Сохранить</GhostButton>
                </form>
                <div className="chats__members">
                  {members.map((m) => (
                    <div key={m.userId} className="chats__member">
                      {m.name} · {m.role}
                    </div>
                  ))}
                </div>
                <p className="chats__hint">Чтобы добавить человека: найдите ник сверху и нажмите «В группу», затем:</p>
                <GhostButton
                  type="button"
                  disabled={groupPick.length === 0}
                  onClick={() => {
                    void Promise.all(groupPick.map((u) => addGroupMember(conversationId, u.id)))
                      .then(() => {
                        setGroupPick([])
                        setToast('Участники добавлены')
                        return reloadMembers(conversationId)
                      })
                      .catch((err) => setError(String(err)))
                  }}
                >
                  Добавить выбранных ({groupPick.length})
                </GhostButton>
              </div>
            )}
            <div className="chats__messages">
              {[...messages].reverse().map((m) => (
                <div key={mid(m)} className={`chats__msg${isOwnMessage(m) ? ' is-own' : ''}`}>
                  {!isOwnMessage(m) && <div className="chats__msgSender">{senderLabel(m)}</div>}
                  {m.attachments?.map((a) =>
                    a.kind === 'image' ? (
                      <button
                        key={a.id}
                        type="button"
                        className="chats__msgImageBtn"
                        onClick={() => setLightbox(mediaURL(a.url))}
                      >
                        <img className="chats__msgImage" src={mediaURL(a.url)} alt={a.filename || 'фото'} />
                      </button>
                    ) : (
                      <a key={a.id} className="chats__msgFile" href={mediaURL(a.url)} target="_blank" rel="noreferrer">
                        📎 {a.filename || 'Файл'}
                      </a>
                    ),
                  )}
                  {m.DeletedForAllAt ? (
                    <div className="chats__msgBody">Сообщение удалено</div>
                  ) : (
                    m.Body && <div className="chats__msgBody">{m.Body}</div>
                  )}
                  <div className="chats__msgMeta">{new Date(m.CreatedAt).toLocaleString()}</div>
                </div>
              ))}
            </div>
            {pendingFiles.length > 0 && (
              <div className="chats__pending">
                {pendingFiles.map((f, i) => (
                  <button
                    key={`${f.name}-${i}`}
                    type="button"
                    className="chats__pendingChip"
                    onClick={() => setPendingFiles((prev) => prev.filter((_, idx) => idx !== i))}
                  >
                    {f.name} ×
                  </button>
                ))}
              </div>
            )}
            <form
              className="chats__composer"
              onSubmit={(e) => {
                e.preventDefault()
                if ((!text.trim() && pendingFiles.length === 0) || !conversationId) return
                const files = pendingFiles
                void sendMessage(conversationId, text, files.length ? files : undefined)
                  .then((m) => {
                    setMessages((prev) => [m, ...prev])
                    setText('')
                    setPendingFiles([])
                    void reloadList()
                  })
                  .catch((err) => setError(String(err)))
              }}
            >
              <input
                ref={fileRef}
                type="file"
                accept="image/*"
                multiple
                hidden
                onChange={(e) => {
                  const files = Array.from(e.target.files || [])
                  if (files.length) setPendingFiles((prev) => [...prev, ...files])
                  e.target.value = ''
                }}
              />
              <GhostButton type="button" className="chats__attach" onClick={() => fileRef.current?.click()}>
                📷
              </GhostButton>
              <GlassInput
                className="chats__composerInput"
                value={text}
                onChange={(e) => onComposerChange(e.target.value)}
                placeholder="Сообщение"
              />
              <PrimaryButton type="submit" className="chats__composerSend">
                Отправить
              </PrimaryButton>
            </form>
          </>
        )}
        {error && <div className="chats__error">{error}</div>}
        {toast && <div className="chats__toast">{toast}</div>}
        {lightbox && (
          <button type="button" className="chats__lightbox" onClick={() => setLightbox(null)}>
            <img src={lightbox} alt="" />
          </button>
        )}
      </section>
    </div>
  )
}
