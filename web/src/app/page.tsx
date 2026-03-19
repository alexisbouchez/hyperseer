'use client'
import { useState, useEffect } from 'react'
import { MOCK_DIGEST } from '@/lib/mock'

function fmt(ms: number): string {
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`
  return `${ms}ms`
}

function StatCard({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div style={{
      flex: 1, minWidth: 120,
      border: '1px solid var(--border)',
      boxShadow: 'var(--shadow)',
      background: 'var(--surface)',
      padding: '12px 14px',
    }}>
      <div style={{ fontSize: 10, letterSpacing: '0.12em', color: 'var(--muted)', marginBottom: 6 }}>{label}</div>
      <div style={{ fontSize: 22, fontWeight: 500, lineHeight: 1, marginBottom: sub ? 4 : 0 }}>{value}</div>
      {sub && <div style={{ fontSize: 11, color: 'var(--muted)', marginTop: 4 }}>{sub}</div>}
    </div>
  )
}

const LEVEL_COLORS: Record<string, { color: string; bg: string }> = {
  TRACE: { color: 'rgba(0,0,0,0.35)', bg: '#f5f5f5' },
  DEBUG: { color: 'rgba(0,0,0,0.35)', bg: '#f5f5f5' },
  INFO:  { color: '#1d4ed8', bg: '#eff6ff' },
  WARN:  { color: '#b45309', bg: '#fffbeb' },
  ERROR: { color: '#b91c1c', bg: '#fef2f2' },
  FATAL: { color: '#7f1d1d', bg: '#fef2f2' },
}

function LevelBadge({ level }: { level: string }) {
  const c = LEVEL_COLORS[level] ?? { color: 'var(--muted)', bg: 'transparent' }
  return (
    <span style={{
      display: 'inline-block',
      fontSize: 10, letterSpacing: '0.06em',
      color: c.color, background: c.bg,
      padding: '1px 5px',
      fontWeight: level === 'FATAL' ? 700 : 400,
      width: 44, textAlign: 'center',
    }}>
      {level}
    </span>
  )
}

function formatTs(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

export default function DigestPage() {
  const [now, setNow] = useState(new Date())
  const d = MOCK_DIGEST

  useEffect(() => {
    const t = setInterval(() => setNow(new Date()), 1000)
    return () => clearInterval(t)
  }, [])

  const ts = now.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })

  return (
    <div style={{ padding: '0', height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* header */}
      <div style={{
        padding: '14px 20px 12px',
        borderBottom: '1px solid var(--border)',
        display: 'flex', alignItems: 'baseline', gap: 12,
      }}>
        <h1 style={{ margin: 0, fontSize: 13, fontWeight: 600, letterSpacing: '0.14em' }}>DIGEST</h1>
        <span style={{ fontSize: 11, color: 'var(--muted)' }}>as of {ts}</span>
        <span style={{ fontSize: 11, color: 'var(--muted)', marginLeft: 'auto' }}>last 24h</span>
      </div>

      <div style={{ flex: 1, overflow: 'auto', padding: 20 }}>
        {/* stats grid */}
        <div style={{ display: 'flex', gap: 8, marginBottom: 24, flexWrap: 'wrap' }}>
          <StatCard label="ERROR RATE" value={`${d.error_rate}%`} sub="of all log events" />
          <StatCard label="REQ / MIN" value={`${d.req_per_min}`} sub="avg over 24h" />
          <StatCard label="P50 LATENCY" value={fmt(d.p50_ms)} />
          <StatCard label="P95 LATENCY" value={fmt(d.p95_ms)} />
          <StatCard label="P99 LATENCY" value={fmt(d.p99_ms)} />
        </div>

        {/* recent errors */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ fontSize: 10, letterSpacing: '0.14em', color: 'var(--muted)', marginBottom: 8 }}>RECENT ERRORS</div>
          <div style={{ border: '1px solid var(--border)', background: 'var(--surface)' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  {['TIME', 'LEVEL', 'SERVICE', 'MESSAGE'].map(h => (
                    <th key={h} style={{
                      textAlign: 'left', padding: '6px 12px',
                      fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)', fontWeight: 400,
                    }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {d.recent_errors.map((log, i) => (
                  <tr key={log.id} style={{
                    borderBottom: i < d.recent_errors.length - 1 ? '1px solid var(--border)' : undefined,
                    background: 'var(--surface)',
                  }}>
                    <td style={{ padding: '5px 12px', color: 'var(--muted)', whiteSpace: 'nowrap', fontSize: 11 }}>
                      {formatTs(log.timestamp)}
                    </td>
                    <td style={{ padding: '5px 12px', whiteSpace: 'nowrap' }}>
                      <LevelBadge level={log.level} />
                    </td>
                    <td style={{ padding: '5px 12px', color: 'var(--muted)', whiteSpace: 'nowrap', fontSize: 11 }}>
                      {log.service_name}
                    </td>
                    <td style={{ padding: '5px 12px', maxWidth: 400, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {log.body}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* service health */}
        <div>
          <div style={{ fontSize: 10, letterSpacing: '0.14em', color: 'var(--muted)', marginBottom: 8 }}>SERVICE HEALTH</div>
          <div style={{ border: '1px solid var(--border)', background: 'var(--surface)' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  {['SERVICE', 'REQUESTS', 'ERRORS', 'ERROR RATE', 'P95'].map(h => (
                    <th key={h} style={{
                      textAlign: 'left', padding: '6px 12px',
                      fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)', fontWeight: 400,
                    }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {d.top_services.map((svc, i) => {
                  const rate = svc.req_count > 0 ? Math.round((svc.error_count / svc.req_count) * 1000) / 10 : 0
                  const isHealthy = rate < 1
                  return (
                    <tr key={svc.name} style={{
                      borderBottom: i < d.top_services.length - 1 ? '1px solid var(--border)' : undefined,
                      background: 'var(--surface)',
                    }}>
                      <td style={{ padding: '5px 12px', fontWeight: 500 }}>{svc.name}</td>
                      <td style={{ padding: '5px 12px', color: 'var(--muted)' }}>{svc.req_count.toLocaleString()}</td>
                      <td style={{ padding: '5px 12px', color: svc.error_count > 0 ? 'var(--red)' : 'var(--muted)' }}>
                        {svc.error_count}
                      </td>
                      <td style={{ padding: '5px 12px' }}>
                        <span style={{ color: isHealthy ? 'var(--green)' : 'var(--red)', fontSize: 11 }}>
                          {rate.toFixed(1)}%
                        </span>
                      </td>
                      <td style={{ padding: '5px 12px', color: 'var(--muted)' }}>{fmt(Math.round(svc.p95_ms))}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
