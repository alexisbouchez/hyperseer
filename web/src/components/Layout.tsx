'use client'
import Link from 'next/link'
import { usePathname } from 'next/navigation'

const NAV = [
  { label: 'DIGEST', to: '/' },
  { label: 'LOGS', to: '/logs' },
  { label: 'TRACES', to: '/traces' },
]

export default function Layout({ email, onSignOut, children }: { email: string | null; onSignOut: () => void; children: React.ReactNode }) {
  const pathname = usePathname()

  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden', background: 'var(--bg)' }}>
      <aside style={{
        width: 180, flexShrink: 0, background: '#fff',
        borderRight: '1px solid var(--border)',
        boxShadow: '2px 0 0 rgba(0,0,0,0.04)',
        display: 'flex', flexDirection: 'column', overflow: 'hidden',
      }}>
        <div style={{ padding: '16px 16px 12px', borderBottom: '1px solid var(--border)' }}>
          <div style={{ fontSize: 11, letterSpacing: '0.18em', fontWeight: 700, color: 'var(--fg)' }}>HYPERSEER</div>
        </div>

        <nav style={{ flex: 1, padding: '8px 0' }}>
          {NAV.map(({ label, to }) => {
            const isActive = to === '/' ? pathname === '/' : pathname.startsWith(to)
            return (
              <Link key={to} href={to} style={{
                display: 'block', padding: '7px 16px',
                fontSize: 11, letterSpacing: '0.12em',
                fontWeight: isActive ? 700 : 400,
                color: isActive ? 'var(--fg)' : 'var(--muted)',
                background: 'transparent',
                borderLeft: isActive ? '2px solid var(--fg)' : '2px solid transparent',
                textDecoration: 'none',
              }}>
                {label}
              </Link>
            )
          })}
        </nav>

        <div style={{ borderTop: '1px solid var(--border)', padding: '12px 16px' }}>
          <div style={{ fontSize: 11, color: 'var(--muted)', marginBottom: 8, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {email ?? '—'}
          </div>
          <button
            onClick={onSignOut}
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

      <main style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', padding: 16 }}>
        <div style={{ background: '#fff', border: '1px solid var(--border)', boxShadow: 'var(--shadow)', minHeight: '100%' }}>
          {children}
        </div>
      </main>
    </div>
  )
}
