'use client'
import { useEffect, useState, Suspense } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { getAuthConfig, exchangeCode, setToken } from '@/lib/auth'

function CallbackInner() {
  const [error, setError] = useState('')
  const router = useRouter()
  const searchParams = useSearchParams()

  useEffect(() => {
    const code = searchParams.get('code')
    if (!code) { setError('no auth code in callback'); return }
    getAuthConfig()
      .then(config => exchangeCode(code, config))
      .then(token => { setToken(token); router.replace('/') })
      .catch(e => setError(String(e)))
  }, [])

  if (error) return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', fontFamily: 'monospace', color: 'var(--red)', padding: 24 }}>
      {error}
    </div>
  )

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', fontFamily: 'monospace', color: 'var(--muted)' }}>
      signing in...
    </div>
  )
}

export default function CallbackPage() {
  return (
    <Suspense fallback={
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', fontFamily: 'monospace', color: 'var(--muted)' }}>
        signing in...
      </div>
    }>
      <CallbackInner />
    </Suspense>
  )
}
