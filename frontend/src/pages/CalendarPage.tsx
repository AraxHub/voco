import { useEffect, useMemo, useState } from 'react'
import { cancelEvent, createEvent, listEvents, rescheduleEvent, type CalendarEvent } from '../lib/api'
import { PrimaryButton, GhostButton } from '../ui/Button'
import { GlassInput } from '../ui/Input'
import './calendar.css'

type View = 'month' | 'week' | 'day'

const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

function startOfDay(d: Date) {
  const x = new Date(d)
  x.setHours(0, 0, 0, 0)
  return x
}

function startOfWeek(d: Date) {
  const x = startOfDay(d)
  const day = (x.getDay() + 6) % 7
  x.setDate(x.getDate() - day)
  return x
}

function sameDay(a: Date, b: Date) {
  return a.toDateString() === b.toDateString()
}

function friendlyError(err: unknown) {
  const msg = (err instanceof Error ? err.message : String(err)).replace(/^Error:\s*/i, '')
  if (/сессия истекла|invalid token|unauthorized|authorization required/i.test(msg)) {
    return 'Сессия истекла или токен недействителен. Войдите снова.'
  }
  if (/validation/i.test(msg) && /назван/i.test(msg)) return msg
  if (/^validation$/i.test(msg)) return 'Укажите название встречи'
  return msg
}

export function CalendarPage() {
  const [view, setView] = useState<View>('week')
  const [cursor, setCursor] = useState(() => startOfDay(new Date()))
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [title, setTitle] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [dragId, setDragId] = useState<string | null>(null)

  const weekStart = useMemo(() => startOfWeek(cursor), [cursor])
  const weekDays = useMemo(
    () => Array.from({ length: 7 }, (_, i) => {
      const d = new Date(weekStart)
      d.setDate(weekStart.getDate() + i)
      return d
    }),
    [weekStart],
  )

  const range = useMemo(() => {
    const from = new Date(cursor)
    const to = new Date(cursor)
    if (view === 'month') {
      from.setDate(1)
      from.setHours(0, 0, 0, 0)
      to.setMonth(to.getMonth() + 1, 0)
      to.setHours(23, 59, 59, 999)
    } else if (view === 'week') {
      from.setTime(weekStart.getTime())
      to.setTime(weekStart.getTime())
      to.setDate(to.getDate() + 7)
      to.setMilliseconds(-1)
    } else {
      from.setHours(0, 0, 0, 0)
      to.setHours(23, 59, 59, 999)
    }
    return { from, to }
  }, [cursor, view, weekStart])

  async function reload() {
    const data = await listEvents(range.from.toISOString(), range.to.toISOString())
    setEvents(Array.isArray(data) ? data : [])
  }

  useEffect(() => {
    void reload().catch((e) => setError(friendlyError(e)))
  }, [range.from.toISOString(), range.to.toISOString()])

  const hours = Array.from({ length: 24 }, (_, i) => i)

  function eventsAt(day: Date, hour: number) {
    return events.filter((ev) => {
      const s = new Date(ev.StartsAt)
      return sameDay(s, day) && s.getHours() === hour
    })
  }

  function onDropSlot(day: Date, hour: number) {
    if (!dragId) return
    const ev = events.find((x) => x.ID === dragId)
    if (!ev) return
    const starts = new Date(day)
    starts.setHours(hour, 0, 0, 0)
    const dur = new Date(ev.EndsAt).getTime() - new Date(ev.StartsAt).getTime()
    const ends = new Date(starts.getTime() + dur)
    void rescheduleEvent(ev.ID, starts.toISOString(), ends.toISOString())
      .then(reload)
      .catch((err) => setError(friendlyError(err)))
    setDragId(null)
  }

  function renderEventCard(ev: CalendarEvent) {
    return (
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
    )
  }

  const navStep = view === 'month' ? 30 : view === 'week' ? 7 : 1
  const heading =
    view === 'week'
      ? `${weekDays[0].toLocaleDateString()} – ${weekDays[6].toLocaleDateString()}`
      : view === 'month'
        ? cursor.toLocaleDateString(undefined, { month: 'long', year: 'numeric' })
        : cursor.toLocaleDateString()

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
          <GhostButton type="button" onClick={() => setCursor(new Date(cursor.getTime() - 86400000 * navStep))}>
            ←
          </GhostButton>
          <strong>{heading}</strong>
          <GhostButton type="button" onClick={() => setCursor(new Date(cursor.getTime() + 86400000 * navStep))}>
            →
          </GhostButton>
        </div>
        <form
          className="cal__create"
          onSubmit={(e) => {
            e.preventDefault()
            const trimmed = title.trim()
            if (!trimmed) {
              setError('Укажите название встречи')
              return
            }
            setError(null)
            const starts = new Date(cursor)
            starts.setHours(11, 0, 0, 0)
            const ends = new Date(starts.getTime() + 60 * 60 * 1000)
            void createEvent({
              title: trimmed,
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
              .catch((err) => setError(friendlyError(err)))
          }}
        >
          <GlassInput value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Новая встреча" />
          <PrimaryButton type="submit">Создать</PrimaryButton>
        </form>
      </header>

      {view === 'month' && (
        <div className="cal__month">
          {WEEKDAYS.map((d) => (
            <div key={d} className="cal__monthHead">
              {d}
            </div>
          ))}
          {Array.from({ length: 42 }, (_, i) => {
            const first = new Date(cursor.getFullYear(), cursor.getMonth(), 1)
            const startPad = (first.getDay() + 6) % 7
            const day = new Date(first)
            day.setDate(i - startPad + 1)
            const inMonth = day.getMonth() === cursor.getMonth()
            const dayEvents = events.filter((ev) => sameDay(new Date(ev.StartsAt), day))
            return (
              <button
                key={i}
                type="button"
                className={`cal__dayCell${inMonth ? '' : ' is-out'}${sameDay(day, new Date()) ? ' is-today' : ''}`}
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

      {view === 'week' && (
        <div className="cal__week">
          <div className="cal__weekCorner" />
          {weekDays.map((d, i) => (
            <button
              key={i}
              type="button"
              className={`cal__weekHead${sameDay(d, new Date()) ? ' is-today' : ''}`}
              onClick={() => {
                setCursor(startOfDay(d))
                setView('day')
              }}
            >
              <span>{WEEKDAYS[i]}</span>
              <strong>{d.getDate()}</strong>
            </button>
          ))}
          {hours.map((h) => (
            <div key={h} className="cal__weekHourRow">
              <div className="cal__hour">{`${h}:00`}</div>
              {weekDays.map((d, di) => (
                <div
                  key={`${h}-${di}`}
                  className={`cal__weekCell${(di + h) % 2 === 0 ? ' is-a' : ' is-b'}`}
                  onDragOver={(e) => e.preventDefault()}
                  onDrop={() => onDropSlot(d, h)}
                >
                  {eventsAt(d, h).map(renderEventCard)}
                </div>
              ))}
            </div>
          ))}
        </div>
      )}

      {view === 'day' && (
        <div className="cal__day">
          {hours.map((h) => (
            <div key={h} className="cal__row">
              <div className="cal__hour">{`${h}:00`}</div>
              <div
                className="cal__slot"
                onDragOver={(e) => e.preventDefault()}
                onDrop={() => onDropSlot(cursor, h)}
              >
                {eventsAt(cursor, h).map(renderEventCard)}
              </div>
            </div>
          ))}
        </div>
      )}

      {error && <div className="cal__error">{error}</div>}
    </div>
  )
}
