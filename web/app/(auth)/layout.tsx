// web/app/(auth)/layout.tsx
import AuthBranding from '@/components/auth/AuthBranding'

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex" style={{ backgroundColor: 'var(--auth-bg-form)' }}>
      {/* Left panel — 40%, hidden on mobile */}
      <div className="lg:w-[40%]">
        <AuthBranding />
      </div>

      {/* Right panel — 60% (full width on mobile) */}
      <div className="flex-1 flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-[400px]">
          {children}
        </div>
      </div>
    </div>
  )
}
