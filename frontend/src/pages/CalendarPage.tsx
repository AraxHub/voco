import { useEffect, useMemo, useState } from 'react'
import { cancelEvent, createEvent, listEvents, rescheduleEvent, type CalendarEvent } from '../lib/api'
import { PrimaryButton, GhostButton } from '../ui/Button'
import { GlassInput } from '../ui/Input'
import './calendar.css'

type View = 'month' | 'week' | 'day'

function startOfDay(d: Date) {
  const x = new Date(d)
  x.setHours(0, 0, 0, 0)
  return x
}

export function CalendarPage() {
  const [view, setView] = useState<View>('week')
  const [cursor, setCursor] = useState(() => startOfDay(new Date()))
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [title, setTitle] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [dragId, setDragId] = useState<string | null>(null)

  const range = useMemo(() => {
    const from = new Date(cursor)
    const to = new Date(cursor)
    if (view === 'month') {
      from.setDate(1)
      to.setMonth(to.getMonth() + 1, 0)
      to.setHours(23, 59, 59, 999)
    } else if (view === 'week') {
      const day = (from.getDay() + 6) % 7
      from.setDate(from.getDate() - day)
      to.setTime(from.getTime())
      to.setDate(to.getDate() + 7)
    } else {
      to.setHours(23, 59, 59, 999)
    }
    return { from, to }
  }, [cursor, view])

  async function reload() {
    const data = await listEvents(range.from.toISOString(), range.to.toISOString())
    setEvents(Array.isArray(data) ? data : [])
  }

  useEffect(() => {
    void reload().catch((e) => setError(String(e)))
  }, [range.from.toISOString(), range.to.toISOString()])

  const hours = Array.from({ length: 24 }, (_, i) => i)

  return (
    <div className="cal">
      <header className="cal__toolbar">
        <div className="cal__views">
          {(['month', 'week', 'day'] as View[]).map((v) => (
            <button key={v} type="button" className={view === v ? 'is-active' : ''} onClick={() => setView(v)}>
              {v === 'month' ? 'Месяц' : v === 'week' ? 'Неделя' : 'День'}
            </button>
          ))}
        </div>
        <div className="cal__nav">
          <GhostButton onClick={() => setCursor(new Date(cursor.getTime() - 86400000 * (view === 'month' ? 30 : view === 'week' ? 7 : 1)))}>
            ←
          </GhostButton>
          <strong>{cursor.toLocaleDateString()}</strong>
          <GhostButton onClick={() => setCursor(new Date(cursor.getTime() + 86400000 * (view === 'month' ? 30 : view === 'week' ? 7 : 1)))}>
            →
          </GhostButton>
        </div>
        <form
          className="cal__create"
          onSubmit={(e) => {
            e.preventDefault()
            const starts = new Date(cursor)
            starts.setHours(11, 0, 0, 0)
            const ends = new Date(starts.getTime() + 60 * 60 * 1000)
            void createEvent({
              title,
              description: '',
              timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
              startsAt: starts.toISOString(),
              endsAt: ends.toISOString(),
              attendeeIds: [],
              reminders: [15],
            })
              .then(() => {
                setTitle('')
                return reload()
              })
              .catch((err) => setError(String(err)))
          }}
        >
          <GlassInput value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Новая встреча" />
          <PrimaryButton type="submit">Создать</PrimaryButton>
        </form>
      </header>

      {view !== 'month' ? (
        <div className="cal__grid">
          {hours.map((h) => (
            <div key={h} className="cal__row">
              <div className="cal__hour">{`${h}:00`}</div>
              <div
                className="cal__slot"
                onDragOver={(e) => e.preventDefault()}
                onDrop={() => {
                  if (!dragId) return
                  const ev = events.find((x) => x.ID === dragId)
                  if (!ev) return
                  const starts = new Date(cursor)
                  starts.setHours(h, 0, 0, 0)
                  const dur = new Date(ev.EndsAt).getTime() - new Date(ev.StartsAt).getTime()
                  const ends = new Date(starts.getTime() + dur)
                  void rescheduleEvent(ev.ID, starts.toISOString(), ends.toISOString())
                    .then(reload)
                    .catch((err) => setError(String(err)))
                  setDragId(null)
                }}
              >
                {events
                  .filter((ev) => {
                    const s = new Date(ev.StartsAt)
                    return s.toDateString() === cursor.toDateString() && s.getHours() === h
                  })
                  .map((ev) => (
                    <div
                      key={ev.ID}
                      className={`cal__event${ev.Status === 'cancelled' ? ' is-cancelled' : ''}`}
                      draggable
                      onDragStart={() => setDragId(ev.ID)}
                    >
                      <div>{ev.Title}</div>
                      <div className="cal__eventMeta">
                        {new Date(ev.StartsAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        {' – '}
                        {new Date(ev.EndsAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        {ev.Status === 'cancelled' ? ' · отменено' : ''}
                      </div>
                      <div className="cal__eventActions">
                        <button
                          type="button"
                          onClick={() => {
                            const s = new Date(ev.StartsAt)
                            const e = new Date(ev.EndsAt)
                            e.setTime(e.getTime() + 30 * 60 * 1000)
                            void rescheduleEvent(ev.ID, s.toISOString(), e.toISOString()).then(reload)
                          }}
                        >
                          +30м
                        </button>
                        <button type="button" onClick={() => void cancelEvent(ev.ID).then(reload)}>
                          Отменить
                        </button>
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="cal__month">
          {Array.from({ length: 42 }, (_, i) => {
            const first = new Date(cursor.getFullYear(), cursor.getMonth(), 1)
            const startPad = (first.getDay() + 6) % 7
            const day = new Date(first)
            day.setDate(i - startPad + 1)
            const dayEvents = events.filter((ev) => new Date(ev.StartsAt).toDateString() === day.toDateString())
            return (
              <button
                key={i}
                type="button"
                className="cal__dayCell"
                onClick={() => {
                  setCursor(startOfDay(day))
                  setView('day')
                }}
              >
                <span>{day.getDate()}</span>
                {dayEvents.slice(0, 3).map((ev) => (
                  <em key={ev.ID}>{ev.Title}</em>
                ))}
              </button>
            )
          })}
        </div>
      )}
      {error && <div className="cal__error">{error}</div>}
    </div>
  )
}
