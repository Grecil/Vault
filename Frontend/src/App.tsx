import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import LandingPage from './components/LandingPage'
import Dashboard from './components/Dashboard'
import ErrorBoundary from './components/ErrorBoundary'

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/dashboard" element={
          <ErrorBoundary>
            <Dashboard />
          </ErrorBoundary>
        } />
      </Routes>
    </Router>
  )
}

export default App
