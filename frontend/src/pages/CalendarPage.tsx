import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  cancelEvent,
  createEvent,
  listEvents,
  rescheduleEvent,
  searchUsers,
  type CalendarAttendee,
  type CalendarEvent,
  type VocoUser,
} from '../lib/api'
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
  if (/отменённую встречу|отмененную встречу/i.test(msg)) return 'Отменённую встречу нельзя переносить'
  if (/validation/i.test(msg) && /назван/i.test(msg)) return msg
  if (/^validation$/i.test(msg)) return 'Укажите название встречи'
  return msg
}

function eventAttendees(ev: CalendarEvent): CalendarAttendee[] {
  return ev.attendees || ev.Attendees || []
}

function eventRoomId(ev: CalendarEvent): string {
  return (ev.roomId || ev.RoomID || '').toString()
}

function toLocalInputValue(d: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function CalendarPage() {
  const nav = useNavigate()
  const [view, setView] = useState<View>('week')
  const [cursor, setCursor] = useState(() => startOfDay(new Date()))
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [error, setError] = useState<string | null>(null)
  const [dragId, setDragId] = useState<string | null>(null)
  const [selected, setSelected] = useState<CalendarEvent | null>(null)
  const [draft, setDraft] = useState<{ starts: Date; ends: Date } | null>(null)
  const [draftTitle, setDraftTitle] = useState('')
  const [draftDesc, setDraftDesc] = useState('')
  const [attendeeQ, setAttendeeQ] = useState('')
  const [attendeeHits, setAttendeeHits] = useState<VocoUser[]>([])
  const [draftAttendees, setDraftAttendees] = useState<VocoUser[]>([])

  const weekStart = useMemo(() => startOfWeek(cursor), [cursor])
  const weekDays = useMemo(
    () =>
      Array.from({ length: 7 }, (_, i) => {
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

  function openCreateAt(day: Date, hour: number) {
    const starts = new Date(day)
    starts.setHours(hour, 0, 0, 0)
    const ends = new Date(starts.getTime() + 60 * 60 * 1000)
    setSelected(null)
    setDraft({ starts, ends })
    setDraftTitle('')
    setDraftDesc('')
    setDraftAttendees([])
    setAttendeeQ('')
    setAttendeeHits([])
  }

  function onDropSlot(day: Date, hour: number) {
    if (!dragId) return
    const ev = events.find((x) => x.ID === dragId)
    if (!ev || ev.Status === 'cancelled') {
      setDragId(null)
      return
    }
    const starts = new Date(day)
    starts.setHours(hour, 0, 0, 0)
    const dur = new Date(ev.EndsAt).getTime() - new Date(ev.StartsAt).getTime()
    const ends = new Date(starts.getTime() + dur)
    void rescheduleEvent(ev.ID, starts.toISOString(), ends.toISOString())
      .then(reload)
      .catch((err) => setError(friendlyError(err)))
    setDragId(null)
  }

  async function searchAttendees(value: string) {
    setAttendeeQ(value)
    if (value.trim().length < 1) {
      setAttendeeHits([])
      return
    }
    try {
      setAttendeeHits(await searchUsers(value.trim()))
    } catch (e) {
      setError(friendlyError(e))
    }
  }

  function renderEventCard(ev: CalendarEvent) {
    const cancelled = ev.Status === 'cancelled'
    return (
      <div
        key={ev.ID}
        className={`cal__event${cancelled ? ' is-cancelled' : ''}`}
        draggable={!cancelled}
        onDragStart={(e) => {
          if (cancelled) {
            e.preventDefault()
            return
          }
          setDragId(ev.ID)
        }}
        onClick={(e) => {
          e.stopPropagation()
          setDraft(null)
          setSelected(ev)
        }}
      >
        <div>{ev.Title}</div>
        <div className="cal__eventMeta">
          {new Date(ev.StartsAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          {' – '}
          {new Date(ev.EndsAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          {cancelled ? ' · отменено' : ''}
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
        <p className="cal__hint">Клик по часу — создать встречу. Клик по встрече — детали.</p>
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
                  <em
                    key={ev.ID}
                    onClick={(e) => {
                      e.stopPropagation()
                      setSelected(ev)
                    }}
                  >
                    {ev.Title}
                  </em>
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
                  onClick={() => openCreateAt(d, h)}
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
                onClick={() => openCreateAt(cursor, h)}
              >
                {eventsAt(cursor, h).map(renderEventCard)}
              </div>
            </div>
          ))}
        </div>
      )}

      {draft && (
        <div className="cal__modal" role="dialog" aria-modal="true">
          <form
            className="cal__modalCard"
            onSubmit={(e) => {
              e.preventDefault()
              const trimmed = draftTitle.trim()
              if (!trimmed) {
                setError('Укажите название встречи')
                return
              }
              setError(null)
              void createEvent({
                title: trimmed,
                description: draftDesc.trim(),
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
                startsAt: draft.starts.toISOString(),
                endsAt: draft.ends.toISOString(),
                attendeeIds: draftAttendees.map((u) => u.id),
                reminders: [15],
              })
                .then(() => {
                  setDraft(null)
                  return reload()
                })
                .catch((err) => setError(friendlyError(err)))
            }}
          >
            <h3>Новая встреча</h3>
            <GlassInput value={draftTitle} onChange={(e) => setDraftTitle(e.target.value)} placeholder="Название" />
            <GlassInput value={draftDesc} onChange={(e) => setDraftDesc(e.target.value)} placeholder="Описание" />
            <label className="cal__field">
              Начало
              <input
                type="datetime-local"
                value={toLocalInputValue(draft.starts)}
                onChange={(e) => {
                  const starts = new Date(e.target.value)
                  if (Number.isNaN(starts.getTime())) return
                  setDraft((prev) => (prev ? { ...prev, starts } : prev))
                }}
              />
            </label>
            <label className="cal__field">
              Конец
              <input
                type="datetime-local"
                value={toLocalInputValue(draft.ends)}
                onChange={(e) => {
                  const ends = new Date(e.target.value)
                  if (Number.isNaN(ends.getTime())) return
                  setDraft((prev) => (prev ? { ...prev, ends } : prev))
                }}
              />
            </label>
            <GlassInput
              value={attendeeQ}
              onChange={(e) => void searchAttendees(e.target.value)}
              placeholder="Пригласить по нику"
            />
            {attendeeHits.length > 0 && (
              <div className="cal__chips">
                {attendeeHits.map((u) => (
                  <button
                    key={u.id}
                    type="button"
                    className="cal__chip"
                    onClick={() => {
                      setDraftAttendees((prev) => (prev.some((x) => x.id === u.id) ? prev : [...prev, u]))
                      setAttendeeHits([])
                      setAttendeeQ('')
                    }}
                  >
                    + {u.displayName || u.nickname || u.email}
                  </button>
                ))}
              </div>
            )}
            {draftAttendees.length > 0 && (
              <div className="cal__chips">
                {draftAttendees.map((u) => (
                  <button
                    key={u.id}
                    type="button"
                    className="cal__chip is-selected"
                    onClick={() => setDraftAttendees((prev) => prev.filter((x) => x.id !== u.id))}
                  >
                    {u.displayName || u.nickname} ×
                  </button>
                ))}
              </div>
            )}
            <div className="cal__modalActions">
              <GhostButton type="button" onClick={() => setDraft(null)}>
                Отмена
              </GhostButton>
              <PrimaryButton type="submit">Создать</PrimaryButton>
            </div>
          </form>
        </div>
      )}

      {selected && (
        <div className="cal__modal" role="dialog" aria-modal="true" onClick={() => setSelected(null)}>
          <div className="cal__modalCard" onClick={(e) => e.stopPropagation()}>
            <h3>{selected.Title}</h3>
            <p className="cal__detailLine">
              {new Date(selected.StartsAt).toLocaleString()} — {new Date(selected.EndsAt).toLocaleString()}
            </p>
            <p className="cal__detailLine">Статус: {selected.Status === 'cancelled' ? 'отменена' : 'запланирована'}</p>
            {selected.Description && <p className="cal__detailLine">{selected.Description}</p>}
            <div className="cal__detailBlock">
              <strong>Приглашённые</strong>
              <ul>
                {eventAttendees(selected).length === 0 && <li>Нет участников</li>}
                {eventAttendees(selected).map((a) => (
                  <li key={a.userId}>
                    {a.name} · {a.status}
                  </li>
                ))}
              </ul>
            </div>
            {eventRoomId(selected) && (
              <div className="cal__detailBlock">
                <strong>Ссылка на комнату</strong>
                <button type="button" className="cal__linkBtn" onClick={() => nav(`/room/${eventRoomId(selected)}`)}>
                  /room/{eventRoomId(selected)}
                </button>
              </div>
            )}
            <div className="cal__modalActions">
              <GhostButton type="button" onClick={() => setSelected(null)}>
                Закрыть
              </GhostButton>
              {selected.Status !== 'cancelled' && (
                <>
                  <GhostButton
                    type="button"
                    onClick={() => {
                      const s = new Date(selected.StartsAt)
                      const e = new Date(selected.EndsAt)
                      e.setTime(e.getTime() + 30 * 60 * 1000)
                      void rescheduleEvent(selected.ID, s.toISOString(), e.toISOString())
                        .then(async () => {
                          await reload()
                          setSelected(null)
                        })
                        .catch((err) => setError(friendlyError(err)))
                    }}
                  >
                    +30м
                  </GhostButton>
                  <PrimaryButton
                    type="button"
                    onClick={() =>
                      void cancelEvent(selected.ID)
                        .then(async () => {
                          await reload()
                          setSelected(null)
                        })
                        .catch((err) => setError(friendlyError(err)))
                    }
                  >
                    Отменить встречу
                  </PrimaryButton>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {error && <div className="cal__error">{error}</div>}
    </div>
  )
}
