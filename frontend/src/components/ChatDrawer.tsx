import { useChat } from '@livekit/components-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import './chatDrawer.css'

type Props = {
  open: boolean
  onClose: () => void
}

function IconArrow() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        d="M14.5 6 8.5 12l6 6"
        stroke="currentColor"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

export function ChatDrawer({ open, onClose }: Props) {
  const { chatMessages, send, isSending } = useChat()
  const [text, setText] = useState('')
  const [sendError, setSendError] = useState<string | null>(null)
  const listRef = useRef<HTMLUListElement | null>(null)

  const items = useMemo(() => chatMessages.slice(-200), [chatMessages])
  const bottomUpItems = useMemo(() => items.slice().reverse(), [items])

  useEffect(() => {
    if (!open) return
    requestAnimationFrame(() => {
      const el = listRef.current
      if (!el) return
      el.scrollTop = el.scrollHeight
    })
  }, [open, items.length])

  return (
    <aside className={`chatDrawer ${open ? 'is-open' : ''}`} aria-hidden={!open}>
      <header className="chatDrawer__header">
        <button type="button" className="chatDrawer__close" onClick={onClose} aria-label="Закрыть чат">
          <IconArrow />
        </button>
        <div className="chatDrawer__title">Messages</div>
      </header>

      <ul className="chatDrawer__list" ref={listRef}>
        {bottomUpItems.map((m, idx) => {
          const name = m.from?.name || m.from?.identity || 'Unknown'
          const mine = Boolean(m.from?.isLocal)
          return (
            <li key={`${m.timestamp}-${idx}`} className={`chatMsg ${mine ? 'chatMsg--mine' : ''}`}>
              <div className="chatMsg__meta">{name}</div>
              <div className="chatMsg__body">{m.message}</div>
            </li>
          )
        })}
      </ul>

      <form
        className="chatDrawer__form"
        onSubmit={async (e) => {
          e.preventDefault()
          const msg = text.trim()
          if (!msg) return
          setSendError(null)
          setText('')
          try {
            await send(msg)
          } catch (err) {
            const m = err instanceof Error ? err.message : String(err)
            setSendError(m)
            console.error('[voco chat] send failed:', err)
          }
        }}
      >
        <input
          className="chatDrawer__input"
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Напиши сообщение…"
        />
        <button className="chatDrawer__send" type="submit" disabled={isSending}>
          Send
        </button>
        {sendError && <p className="chatDrawer__error">{sendError}</p>}
      </form>
    </aside>
  )
}

