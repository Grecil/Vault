import React from 'react'
import { 
  SignedIn, 
  SignedOut, 
  SignInButton, 
  SignUpButton 
} from '@clerk/clerk-react'

interface HeroSectionProps {
  goToDashboard: () => void
}

const HeroSection: React.FC<HeroSectionProps> = ({ goToDashboard }) => {
  return (
    <section className="pt-12 sm:pt-16 lg:pt-20 pb-20 sm:pb-24 lg:pb-32 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto text-center">
        <h1 className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-bold text-foreground mb-6 sm:mb-8">
          Simple File Storage
          <span className="block text-primary">Made Easy</span>
        </h1>
        <p className="text-lg sm:text-xl text-muted-foreground mb-8 sm:mb-12 max-w-3xl mx-auto leading-relaxed">
          Store, organize, and access your files from anywhere. 
          A straightforward file storage solution that just works.
        </p>
        <div className="flex flex-col sm:flex-row gap-3 sm:gap-4 justify-center max-w-md sm:max-w-none mx-auto">
          <SignedIn>
            <button 
              onClick={goToDashboard}
              className="bg-primary text-primary-foreground px-6 sm:px-8 py-3 sm:py-4 rounded-lg text-base sm:text-lg font-semibold hover:opacity-90 transition-all transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
            >
              Go to Dashboard
            </button>
          </SignedIn>
          
          <SignedOut>
            <SignUpButton mode="modal">
              <button className="bg-primary text-primary-foreground px-6 sm:px-8 py-3 sm:py-4 rounded-lg text-base sm:text-lg font-semibold hover:opacity-90 transition-all transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background">
                Sign Up
              </button>
            </SignUpButton>
            <SignInButton mode="modal">
              <button className="border-2 border-border text-foreground px-6 sm:px-8 py-3 sm:py-4 rounded-lg text-base sm:text-lg font-semibold hover:bg-muted transition-all focus:outline-none focus:ring-2 focus:ring-border focus:ring-offset-2 focus:ring-offset-background">
                Login
              </button>
            </SignInButton>
          </SignedOut>
        </div>
      </div>
    </section>
  )
}

export default HeroSection
