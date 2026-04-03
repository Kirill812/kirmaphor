// web/components/auth/AuthBranding.tsx
export default function AuthBranding() {
  return (
    <div
      className="relative hidden lg:flex flex-col justify-between h-full px-10 py-10 overflow-hidden"
      style={{ backgroundColor: 'var(--auth-bg-deep)' }}
    >
      {/* Glow blob — centered */}
      <div
        className="absolute inset-0 flex items-center justify-center pointer-events-none"
        aria-hidden="true"
      >
        <div
          className="w-[500px] h-[500px] rounded-full animate-glow-pulse"
          style={{
            background: 'radial-gradient(circle, var(--auth-glow) 0%, transparent 70%)',
          }}
        />
      </div>

      {/* Logo — top-left */}
      <div className="relative z-10">
        <span className="text-xl font-bold text-white tracking-tight">Kirmaphore</span>
      </div>

      {/* Tagline — bottom-left */}
      <div className="relative z-10">
        <p className="text-sm text-gray-400">Your infrastructure understands you.</p>
      </div>
    </div>
  )
}
