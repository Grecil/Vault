import React from 'react'
import { 
  SignedIn, 
  SignedOut, 
  SignUpButton 
} from '@clerk/clerk-react'

interface CTASectionProps {
  goToDashboard: () => void
}

const CTASection: React.FC<CTASectionProps> = ({ goToDashboard }) => {
  return (
    <section className="py-16 sm:py-20 px-4 sm:px-6 lg:px-8 bg-primary">
      <div className="max-w-4xl mx-auto text-center">
        <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-primary-foreground mb-4">
          Ready to organize your files?
        </h2>
        <p className="text-lg sm:text-xl text-primary-foreground/80 mb-6 sm:mb-8 leading-relaxed">
          Start managing your files the easy way. Simple, convenient, and hassle-free.
        </p>
        <SignedIn>
          <button 
            onClick={goToDashboard}
            className="bg-background text-foreground px-6 sm:px-8 py-3 sm:py-4 rounded-lg text-base sm:text-lg font-semibold hover:bg-card transition-all transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-background focus:ring-offset-2 focus:ring-offset-primary"
          >
            Go to Dashboard
          </button>
        </SignedIn>
        
        <SignedOut>
          <SignUpButton mode="modal">
            <button className="bg-background text-foreground px-6 sm:px-8 py-3 sm:py-4 rounded-lg text-base sm:text-lg font-semibold hover:bg-card transition-all transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-background focus:ring-offset-2 focus:ring-offset-primary">
              Sign Up Now
            </button>
          </SignUpButton>
        </SignedOut>
      </div>
    </section>
  )
}

export default CTASection
