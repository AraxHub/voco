import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  acceptRequest,
  blockRequest,
  callFromChat,
  createGroup,
  listConversations,
  listMessages,
  openDirect,
  searchUsers,
  sendMessage,
  type Conversation,
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

export function ChatsPage() {
  const auth = useAuth()
  const nav = useNavigate()
  const { conversationId } = useParams()
  const [list, setList] = useState<Conversation[]>([])
  const [messages, setMessages] = useState<Message[]>([])
  const [q, setQ] = useState('')
  const [found, setFound] = useState<VocoUser[]>([])
  const [text, setText] = useState('')
  const [groupTitle, setGroupTitle] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)

  const active = useMemo(
    () => list.find((c) => cid(c) === conversationId) || null,
    [list, conversationId],
  )

  async function reloadList() {
    const data = await listConversations()
    setList(Array.isArray(data) ? data : [])
  }

  useEffect(() => {
    void reloadList().catch((e) => setError(String(e)))
  }, [])

  useEffect(() => {
    if (!conversationId) {
      setMessages([])
      return
    }
    void listMessages(conversationId)
      .then((m) => setMessages(Array.isArray(m) ? m : []))
      .catch((e) => setError(String(e)))
  }, [conversationId])

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
          if (conversationId) void listMessages(conversationId).then(setMessages)
          setToast('Новое сообщение')
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
              <button
                key={u.id}
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
            ))}
          </div>
        )}
        <div className="chats__newGroup">
          <GlassInput value={groupTitle} onChange={(e) => setGroupTitle(e.target.value)} placeholder="Название группы" />
          <GhostButton
            type="button"
            onClick={() => {
              const trimmed = groupTitle.trim()
              if (!trimmed) {
                setError('Укажите название группы')
                return
              }
              setError(null)
              void createGroup(trimmed, [])
                .then((c) => {
                  void reloadList()
                  nav(`/chats/${cid(c)}`)
                  setGroupTitle('')
                })
                .catch((e) => {
                  const msg = e instanceof Error ? e.message : String(e)
                  setError(/validation/i.test(msg) ? 'Укажите название группы' : msg)
                })
            }}
          >
            + Группа
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
                <GhostButton
                  onClick={() =>
                    void acceptRequest(cid(active))
                      .then(() => setToast('Запрос принят'))
                      .catch((e) => setError(String(e)))
                  }
                >
                  Принять
                </GhostButton>
                <GhostButton
                  onClick={() =>
                    void blockRequest(cid(active))
                      .then(() => setToast('Заблокировано'))
                      .catch((e) => setError(String(e)))
                  }
                >
                  Заблокировать
                </GhostButton>
                <PrimaryButton
                  onClick={() =>
                    void callFromChat(cid(active))
                      .then((r) => nav(`/room/${r.roomId}`))
                      .catch((e) => setError(String(e)))
                  }
                >
                  Позвонить
                </PrimaryButton>
              </div>
            </header>
            <div className="chats__messages">
              {[...messages].reverse().map((m) => (
                <div key={mid(m)} className="chats__msg">
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
                  })
                  .catch((err) => setError(String(err)))
              }}
            >
              <GlassInput value={text} onChange={(e) => setText(e.target.value)} placeholder="Сообщение" />
              <PrimaryButton type="submit">Отправить</PrimaryButton>
            </form>
          </>
        )}
        {error && <div className="chats__error">{error}</div>}
        {toast && <div className="chats__toast">{toast}</div>}
      </section>
    </div>
  )
}
