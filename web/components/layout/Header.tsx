'use client'

import { useAuth } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export function Header() {
  const { user, logout } = useAuth()
  const router = useRouter()

  const handleLogout = () => {
    logout()
    router.push('/login')
  }

  return (
    <header className="h-14 border-b flex items-center justify-end px-6">
      <DropdownMenu>
        <DropdownMenuTrigger className="flex items-center gap-2 text-sm outline-none">
          <Avatar className="h-8 w-8">
            <AvatarFallback>
              {user?.display_name?.[0]?.toUpperCase() ?? '?'}
            </AvatarFallback>
          </Avatar>
          <span className="hidden sm:block">{user?.display_name}</span>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={() => router.push('/settings/profile')}>
            Profile
          </DropdownMenuItem>
          <DropdownMenuItem onClick={handleLogout} className="text-destructive">
            Sign out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
