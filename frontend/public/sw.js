/* Voco Web Push service worker */
self.addEventListener('push', (event) => {
  let title = 'Voco'
  let body = ''
  try {
    const data = event.data ? event.data.json() : {}
    title = data.title || title
    body = data.body || ''
  } catch {
    body = event.data ? event.data.text() : ''
  }
  event.waitUntil(self.registration.showNotification(title, { body, icon: '/favicon.svg' }))
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  event.waitUntil(clients.openWindow('/'))
})
