// web/middleware.ts
import { NextRequest, NextResponse } from 'next/server'

export function middleware(req: NextRequest) {
  const token = req.cookies.get('kirmaphore_token')?.value
  const path = req.nextUrl.pathname

  const isPublic =
    path === '/' ||
    path === '/login' ||
    path === '/register' ||
    path.startsWith('/login/') ||
    path.startsWith('/register/')

  if (!token && !isPublic) {
    return NextResponse.redirect(new URL('/login', req.url))
  }
  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
}
