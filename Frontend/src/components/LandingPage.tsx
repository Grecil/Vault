import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useDarkMode } from '../hooks/useDarkMode'
import NavSection from './landing/NavSection'
import HeroSection from './landing/HeroSection'
import FeaturesSection from './landing/FeaturesSection'
import CTASection from './landing/CTASection'
import FooterSection from './landing/FooterSection'

const LandingPage = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const { isDarkMode, toggleDarkMode } = useDarkMode()
  const navigate = useNavigate()

  const goToDashboard = () => {
    navigate('/dashboard')
  }

  return (
    <div className="min-h-screen bg-background">
      <NavSection 
        isMenuOpen={isMenuOpen}
        setIsMenuOpen={setIsMenuOpen}
        isDarkMode={isDarkMode}
        toggleDarkMode={toggleDarkMode}
      />
      <HeroSection goToDashboard={goToDashboard} />
      <FeaturesSection />
      <CTASection goToDashboard={goToDashboard} />
      <FooterSection />
    </div>
  )
}

export default LandingPage
