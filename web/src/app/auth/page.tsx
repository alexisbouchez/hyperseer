'use client'
import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { supabase } from '@/lib/supabase'

export default function AuthPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const router = useRouter()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    const { error } = await supabase.auth.signInWithPassword({ email, password })
    if (error) setError(error.message)
    else router.push('/')
    setLoading(false)
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'var(--bg)' }}>
      <div style={{ width: 380, background: 'var(--surface)', border: '1px solid var(--border)', padding: '32px 28px' }}>
        <div style={{ marginBottom: 32 }}>
          <div style={{ fontSize: 11, letterSpacing: '0.15em', color: 'var(--muted)', marginBottom: 4 }}>HYPERSEER</div>
          <div style={{ fontSize: 18, fontWeight: 500 }}>Sign in</div>
        </div>

        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: 16 }}>
            <label style={{ display: 'block', fontSize: 11, color: 'var(--muted)', letterSpacing: '0.08em', marginBottom: 6 }}>EMAIL</label>
            <input
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
              style={{
                width: '100%', padding: '8px 10px', border: '1px solid var(--border)',
                background: 'var(--bg)', fontFamily: 'inherit', fontSize: 13, color: 'var(--fg)',
                outline: 'none',
              }}
              onFocus={e => (e.target.style.borderColor = 'var(--fg)')}
              onBlur={e => (e.target.style.borderColor = 'var(--border)')}
            />
          </div>

          <div style={{ marginBottom: 24 }}>
            <label style={{ display: 'block', fontSize: 11, color: 'var(--muted)', letterSpacing: '0.08em', marginBottom: 6 }}>PASSWORD</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              style={{
                width: '100%', padding: '8px 10px', border: '1px solid var(--border)',
                background: 'var(--bg)', fontFamily: 'inherit', fontSize: 13, color: 'var(--fg)',
                outline: 'none',
              }}
              onFocus={e => (e.target.style.borderColor = 'var(--fg)')}
              onBlur={e => (e.target.style.borderColor = 'var(--border)')}
            />
          </div>

          {error && (
            <div style={{ marginBottom: 16, fontSize: 12, color: 'var(--red)', padding: '8px 10px', border: '1px solid var(--red)', background: '#fef2f2' }}>
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            style={{
              width: '100%', padding: '9px', background: 'var(--fg)', color: '#fff',
              border: 'none', fontFamily: 'inherit', fontSize: 13, cursor: loading ? 'wait' : 'pointer',
              letterSpacing: '0.04em', opacity: loading ? 0.7 : 1,
            }}
          >
            {loading ? 'signing in...' : 'SIGN IN'}
          </button>
        </form>
      </div>
    </div>
  )
}
