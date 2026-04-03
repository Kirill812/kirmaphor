// web/app/(marketing)/page.tsx
import LandingNav from '@/components/landing/LandingNav'
import HeroSection from '@/components/landing/HeroSection'
import PainSection from '@/components/landing/PainSection'
import FeaturesBento from '@/components/landing/FeaturesBento'
import HowItWorksSection from '@/components/landing/HowItWorksSection'
import ForgeAISection from '@/components/landing/ForgeAISection'
import ComparisonSection from '@/components/landing/ComparisonSection'
import PricingSection from '@/components/landing/PricingSection'
import FooterCTASection from '@/components/landing/FooterCTASection'
import LandingFooter from '@/components/landing/LandingFooter'
import ScrollReveal from '@/components/landing/ScrollReveal'

export default function LandingPage() {
  return (
    <main style={{ backgroundColor: '#060A14' }}>
      <LandingNav />
      <HeroSection />
      <ScrollReveal><PainSection /></ScrollReveal>
      <ScrollReveal delay={50}><FeaturesBento /></ScrollReveal>
      <ScrollReveal delay={50}><HowItWorksSection /></ScrollReveal>
      <ScrollReveal delay={50}><ForgeAISection /></ScrollReveal>
      <ScrollReveal delay={50}><ComparisonSection /></ScrollReveal>
      <ScrollReveal delay={50}><PricingSection /></ScrollReveal>
      <ScrollReveal delay={50}><FooterCTASection /></ScrollReveal>
      <LandingFooter />
    </main>
  )
}
