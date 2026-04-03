import { Sidebar } from '@/components/layout/Sidebar'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div
      className="flex min-h-screen"
      style={{ backgroundColor: 'var(--dash-bg)', color: 'var(--dash-text)' }}
    >
      <Sidebar />
      <div className="flex-1 flex flex-col min-w-0">
        {children}
      </div>
    </div>
  )
}
