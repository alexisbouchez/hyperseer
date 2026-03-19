'use client'
import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import type { Session } from '@supabase/supabase-js'
import { supabase } from '@/lib/supabase'

const NAV = [
  { label: 'DIGEST', to: '/' },
  { label: 'LOGS', to: '/logs' },
  { label: 'TRACES', to: '/traces' },
]

export default function Layout({ session, children }: { session: Session | null, children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()

  async function signOut() {
    await supabase.auth.signOut()
    router.push('/auth')
  }

  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden' }}>
      {/* sidebar */}
      <aside style={{
        width: 180, flexShrink: 0, background: '#f8f8f8', borderRight: '1px solid var(--border)',
        display: 'flex', flexDirection: 'column', overflow: 'hidden',
      }}>
        <div style={{ padding: '16px 16px 12px', borderBottom: '1px solid var(--border)' }}>
          <div style={{ fontSize: 11, letterSpacing: '0.18em', fontWeight: 600, color: 'var(--fg)' }}>HYPERSEER</div>
        </div>

        <nav style={{ flex: 1, padding: '8px 0' }}>
          {NAV.map(({ label, to }) => {
            const isActive = to === '/' ? pathname === '/' : pathname.startsWith(to)
            return (
              <Link key={to} href={to} style={{
                display: 'block', padding: '7px 16px',
                fontSize: 11, letterSpacing: '0.12em',
                fontWeight: isActive ? 500 : 400,
                color: isActive ? '#fff' : 'var(--muted)',
                background: isActive ? 'var(--fg)' : 'transparent',
                textDecoration: 'none',
              }}>
                {label}
              </Link>
            )
          })}
        </nav>

        <div style={{ borderTop: '1px solid var(--border)', padding: '12px 16px' }}>
          <div style={{ fontSize: 11, color: 'var(--muted)', marginBottom: 8, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {session?.user.email ?? '—'}
          </div>
          <button
            onClick={signOut}
            style={{
              background: 'none', border: '1px solid var(--border)', padding: '4px 10px',
              fontFamily: 'inherit', fontSize: 11, cursor: 'pointer', color: 'var(--muted)',
              letterSpacing: '0.08em',
            }}
          >
            SIGN OUT
          </button>
        </div>
      </aside>

      {/* main */}
      <main style={{ flex: 1, overflow: 'auto', background: 'var(--bg)' }}>
        {children}
      </main>
    </div>
  )
}
