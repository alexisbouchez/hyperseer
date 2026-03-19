'use client'
import { useState } from 'react'
import { getAuthConfig, startLogin } from '@/lib/auth'

export default function AuthPage() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleSignIn() {
    setLoading(true)
    setError('')
    try {
      const config = await getAuthConfig()
      await startLogin(config)
    } catch {
      setError('Could not reach auth server')
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f4f4f3' }}>
      <div style={{ width: 380, background: 'var(--surface)', border: '1px solid var(--border)', padding: '32px 28px' }}>
        <div style={{ marginBottom: 32 }}>
          <div style={{ fontSize: 11, letterSpacing: '0.15em', color: 'var(--muted)', marginBottom: 4 }}>HYPERSEER</div>
          <div style={{ fontSize: 18, fontWeight: 500 }}>Sign in</div>
        </div>

        {error && (
          <div style={{ marginBottom: 16, fontSize: 12, color: 'var(--red)', padding: '8px 10px', border: '1px solid var(--red)' }}>
            {error}
          </div>
        )}

        <button
          onClick={handleSignIn}
          disabled={loading}
          style={{
            width: '100%', padding: '9px', background: 'var(--fg)', color: '#fff',
            border: 'none', fontFamily: 'inherit', fontSize: 13, cursor: loading ? 'wait' : 'pointer',
            letterSpacing: '0.04em', opacity: loading ? 0.7 : 1,
          }}
        >
          {loading ? 'redirecting...' : 'SIGN IN WITH KEYCLOAK'}
        </button>
      </div>
    </div>
  )
}
