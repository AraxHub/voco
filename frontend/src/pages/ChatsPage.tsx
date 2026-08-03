import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  acceptRequest,
  addGroupMember,
  blockRequest,
  callFromChat,
  createGroup,
  getConversationRequest,
  listConversations,
  listMembers,
  listMessages,
  openDirect,
  renameGroup,
  searchUsers,
  sendMessage,
  type Conversation,
  type ConversationMember,
  type Message,
  type VocoUser,
  wsURL,
} from '../lib/api'
import { useAuth } from '../context/AuthContext'
import { keycloak } from '../lib/keycloak'
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

export function ChatsPage() {
  const auth = useAuth()
  const nav = useNavigate()
  const { conversationId } = useParams()
  const [list, setList] = useState<Conversation[]>([])
  const [messages, setMessages] = useState<Message[]>([])
  const [members, setMembers] = useState<ConversationMember[]>([])
  const [q, setQ] = useState('')
  const [found, setFound] = useState<VocoUser[]>([])
  const [text, setText] = useState('')
  const [groupTitle, setGroupTitle] = useState('')
  const [groupPick, setGroupPick] = useState<VocoUser[]>([])
  const [editTitle, setEditTitle] = useState('')
  const [showGroupEdit, setShowGroupEdit] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)
  const [incomingPending, setIncomingPending] = useState(false)

  const active = useMemo(
    () => list.find((c) => cid(c) === conversationId) || null,
    [list, conversationId],
  )
  const activeIsGroup = active ? typeOf(active) === 'group' : false
  const myId = (keycloak?.tokenParsed?.sub ?? '').toString()
  const myNames = useMemo(() => {
    const names = [auth.username, auth.displayName].filter(Boolean).map((s) => s!.toLowerCase())
    return new Set(names)
  }, [auth.username, auth.displayName])
  function isOwnMessage(m: Message) {
    const sid = senderId(m)
    if (myId && sid && sid === myId) return true
    const name = (m.senderName || '').toLowerCase()
    return Boolean(name && myNames.has(name))
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
    void reloadList().catch((e) => setError(String(e)))
  }, [])

  useEffect(() => {
    if (!conversationId) {
      setMessages([])
      setMembers([])
      setIncomingPending(false)
      setShowGroupEdit(false)
      return
    }
    void listMessages(conversationId)
      .then((m) => setMessages(Array.isArray(m) ? m : []))
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
        const msg = JSON.parse(String(ev.data)) as { event?: string; payload?: { Body?: string; title?: string } }
        if (msg.event === 'message.created') {
          void reloadList()
          if (conversationId) {
            void listMessages(conversationId).then(setMessages)
            void reloadRequest(conversationId)
          }
          setToast('Новое сообщение')
        }
        if (msg.event === 'conversation.updated' && conversationId) {
          void reloadList()
          void reloadMembers(conversationId)
        }
        if (msg.event === 'notification') {
          setToast(msg.payload?.title || 'Уведомление')
        }
      } catch {
        /* ignore */
      }
    }
    return () => ws.close()
  }, [auth.token, conversationId])

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
                  @{u.nickname || u.email}
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
              {titleOf(c)}
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
              <h2>{titleOf(active)}</h2>
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
                  <div className="chats__msgBody">{m.DeletedForAllAt ? 'Сообщение удалено' : m.Body}</div>
                  <div className="chats__msgMeta">{new Date(m.CreatedAt).toLocaleString()}</div>
                </div>
              ))}
            </div>
            <form
              className="chats__composer"
              onSubmit={(e) => {
                e.preventDefault()
                if (!text.trim() || !conversationId) return
                void sendMessage(conversationId, text)
                  .then((m) => {
                    setMessages((prev) => [m, ...prev])
                    setText('')
                    void reloadList()
                  })
                  .catch((err) => setError(String(err)))
              }}
            >
              <GlassInput
                className="chats__composerInput"
                value={text}
                onChange={(e) => setText(e.target.value)}
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
      </section>
    </div>
  )
}
