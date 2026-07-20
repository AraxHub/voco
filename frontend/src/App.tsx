import './App.css'
import { Navigate, Route, Routes } from 'react-router-dom'
import { FuturisticBackground } from './components/FuturisticBackground'
import { ThemeToggle } from './components/ThemeToggle'
import { AccountPage } from './pages/AccountPage'
import { HomePage } from './pages/HomePage'
import { RoomPage } from './pages/RoomPage'

function App() {
  return (
    <>
      <FuturisticBackground />
      <ThemeToggle className="themeToggle--fixed" />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/account" element={<AccountPage />} />
        <Route path="/room/:roomId" element={<RoomPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </>
  )
}

export default App
