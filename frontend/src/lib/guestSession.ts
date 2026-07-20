const guestKey = (roomId: string) => `voco:guest:${roomId}`

export function hasGuestAccess(roomId: string): boolean {
  try {
    return sessionStorage.getItem(guestKey(roomId)) === '1'
  } catch {
    return false
  }
}

export function allowGuestAccess(roomId: string): void {
  try {
    sessionStorage.setItem(guestKey(roomId), '1')
  } catch {
    // ignore
  }
}
