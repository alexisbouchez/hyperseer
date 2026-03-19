'use client'
import { useEffect, useState } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { getToken, getUserEmail, removeToken } from '@/lib/auth'
import Layout from './Layout'

export default function AuthProvider({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(false)
  const [email, setEmail] = useState<string | null>(null)
  const router = useRouter()
  const pathname = usePathname()
  const isAuthRoute = pathname === '/auth' || pathname.startsWith('/auth/')

  useEffect(() => {
    const token = getToken()
    if (token) {
      setEmail(getUserEmail())
    } else if (!isAuthRoute) {
      router.replace('/auth')
    }
    setReady(true)
  }, [])

  if (!ready) return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', fontFamily: 'monospace', color: 'var(--muted)' }}>
      loading...
    </div>
  )

  if (isAuthRoute) return <>{children}</>
  if (!getToken()) return null

  function handleSignOut() {
    removeToken()
    router.replace('/auth')
  }

  return <Layout email={email} onSignOut={handleSignOut}>{children}</Layout>
}
