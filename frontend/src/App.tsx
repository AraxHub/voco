import './App.css'
import { Navigate, Route, Routes } from 'react-router-dom'
import { FuturisticBackground } from './components/FuturisticBackground'
import { ThemeToggle } from './components/ThemeToggle'
import { AppShell, RequireAuth } from './components/AppShell'
import { AccountPage } from './pages/AccountPage'
import { CalendarPage } from './pages/CalendarPage'
import { ChatsPage } from './pages/ChatsPage'
import { HomePage } from './pages/HomePage'
import { RoomPage } from './pages/RoomPage'

function App() {
  return (
    <>
      <FuturisticBackground />
      <ThemeToggle className="themeToggle--fixed" />
      <Routes>
        <Route element={<AppShell />}>
          <Route path="/" element={<HomePage />} />
          <Route
            path="/chats"
            element={
              <RequireAuth>
                <ChatsPage />
              </RequireAuth>
            }
          />
          <Route
            path="/chats/:conversationId"
            element={
              <RequireAuth>
                <ChatsPage />
              </RequireAuth>
            }
          />
          <Route
            path="/calendar"
            element={
              <RequireAuth>
                <CalendarPage />
              </RequireAuth>
            }
          />
          <Route
            path="/account"
            element={
              <RequireAuth>
                <AccountPage />
              </RequireAuth>
            }
          />
        </Route>
        <Route path="/room/:roomId" element={<RoomPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </>
  )
}

export default App
