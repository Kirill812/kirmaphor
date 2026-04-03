// web/components/landing/LandingFooter.tsx
export default function LandingFooter() {
  return (
    <footer
      className="flex flex-col items-center justify-between gap-3 px-6 py-6 text-[12px] sm:flex-row lg:px-12"
      style={{
        borderTop: '1px solid rgba(255,255,255,0.06)',
        backgroundColor: 'var(--land-bg)',
        color: 'rgba(255,255,255,0.22)',
      }}
    >
      <span>Kirmaphore · © 2026</span>
      <div className="flex gap-6">
        <a href="#" className="hover:text-white/50 transition-colors">Docs</a>
        <a href="https://github.com/Kirill812/kirmaphor" target="_blank" rel="noopener noreferrer" className="hover:text-white/50 transition-colors">GitHub</a>
        <a href="#" className="hover:text-white/50 transition-colors">Twitter</a>
      </div>
      <span>Built with ♥ for infrastructure engineers</span>
    </footer>
  )
}
