'use client'
import { useState, useMemo, useRef, useEffect } from 'react'
import { apiFetch } from '@/lib/auth'

interface Log {
  time: string
  severity: string
  service_name: string
  body: string
}

const LEVELS = ['TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL']
const TIME_RANGES = [
  { label: '5M',  ms: 5 * 60 * 1000 },
  { label: '15M', ms: 15 * 60 * 1000 },
  { label: '1H',  ms: 60 * 60 * 1000 },
  { label: '6H',  ms: 6 * 60 * 60 * 1000 },
  { label: '24H', ms: 24 * 60 * 60 * 1000 },
  { label: '7D',  ms: 7 * 24 * 60 * 60 * 1000 },
]

const LEVEL_COLORS: Record<string, { color: string; bg: string }> = {
  TRACE: { color: 'rgba(0,0,0,0.35)', bg: '#f5f5f5' },
  DEBUG: { color: 'rgba(0,0,0,0.35)', bg: '#f5f5f5' },
  INFO:  { color: '#1d4ed8', bg: '#eff6ff' },
  WARN:  { color: '#b45309', bg: '#fffbeb' },
  ERROR: { color: '#b91c1c', bg: '#fef2f2' },
  FATAL: { color: '#7f1d1d', bg: '#fef2f2' },
}

function LevelBadge({ level }: { level: string }) {
  const key = level.toUpperCase()
  const c = LEVEL_COLORS[key] ?? { color: 'var(--muted)', bg: 'transparent' }
  return (
    <span style={{
      display: 'inline-block', fontSize: 10, letterSpacing: '0.05em',
      color: c.color, background: c.bg, padding: '1px 4px',
      fontWeight: key === 'FATAL' ? 700 : 400,
      width: 42, textAlign: 'center', flexShrink: 0,
    }}>
      {key.slice(0, 4)}
    </span>
  )
}

function formatTs(iso: string): string {
  const d = new Date(iso)
  const date = d.toLocaleDateString('en-US', { month: '2-digit', day: '2-digit' })
  const time = d.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
  return `${date} ${time}`
}

export default function LogsPage() {
  const [timeRangeIdx, setTimeRangeIdx] = useState(2)
  const [activeLevels, setActiveLevels] = useState<Set<string>>(new Set(LEVELS))
  const [service, setService] = useState('all')
  const [searchRaw, setSearchRaw] = useState('')
  const [expanded, setExpanded] = useState<number | null>(null)
  const [logs, setLogs] = useState<Log[]>([])
  const [loading, setLoading] = useState(false)
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const searchTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  function handleSearch(v: string) {
    setSearchRaw(v)
    if (searchTimer.current) clearTimeout(searchTimer.current)
    searchTimer.current = setTimeout(() => setDebouncedSearch(v), 150)
  }

  function toggleLevel(level: string) {
    setActiveLevels(prev => {
      const next = new Set(prev)
      if (next.has(level)) next.delete(level)
      else next.add(level)
      return next
    })
  }

  useEffect(() => {
    const to = new Date()
    const from = new Date(to.getTime() - TIME_RANGES[timeRangeIdx].ms)
    const params = new URLSearchParams({ from: from.toISOString(), to: to.toISOString(), limit: '500' })
    setLoading(true)
    apiFetch(`/api/v1/logs?${params}`)
      .then(r => r.ok ? r.json() : Promise.reject(r.status))
      .then(data => setLogs(data ?? []))
      .catch(() => setLogs([]))
      .finally(() => setLoading(false))
  }, [timeRangeIdx])

  const services = useMemo(() => ['all', ...Array.from(new Set(logs.map(l => l.service_name))).sort()], [logs])

  const filtered = useMemo(() => {
    const q = debouncedSearch.toLowerCase()
    return logs.filter(log => {
      if (!activeLevels.has(log.severity.toUpperCase())) return false
      if (service !== 'all' && log.service_name !== service) return false
      if (q && !log.body.toLowerCase().includes(q) && !log.service_name.toLowerCase().includes(q)) return false
      return true
    })
  }, [logs, activeLevels, service, debouncedSearch])

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{
        position: 'sticky', top: 0, zIndex: 10, background: 'var(--bg)',
        borderBottom: '1px solid var(--border)', padding: '10px 16px',
        display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap',
      }}>
        <div style={{ display: 'flex', gap: 0 }}>
          {TIME_RANGES.map((r, i) => (
            <button key={r.label} onClick={() => setTimeRangeIdx(i)} style={{
              padding: '3px 8px', fontSize: 11, letterSpacing: '0.06em',
              border: '1px solid var(--border)', marginLeft: i === 0 ? 0 : -1,
              background: i === timeRangeIdx ? 'var(--fg)' : 'var(--surface)',
              color: i === timeRangeIdx ? '#fff' : 'var(--muted)',
              cursor: 'pointer', fontFamily: 'inherit',
            }}>{r.label}</button>
          ))}
        </div>

        <div style={{ display: 'flex', gap: 4 }}>
          {LEVELS.map(level => {
            const c = LEVEL_COLORS[level]
            const active = activeLevels.has(level)
            return (
              <button key={level} onClick={() => toggleLevel(level)} style={{
                padding: '2px 7px', fontSize: 10, letterSpacing: '0.06em',
                border: `1px solid ${active ? c.color : 'var(--border)'}`,
                background: active ? c.bg : 'var(--surface)',
                color: active ? c.color : 'var(--muted)',
                cursor: 'pointer', fontFamily: 'inherit',
                fontWeight: level === 'FATAL' && active ? 700 : 400,
                opacity: active ? 1 : 0.5,
              }}>{level}</button>
            )
          })}
        </div>

        <select value={service} onChange={e => setService(e.target.value)} style={{
          padding: '3px 8px', fontSize: 11, border: '1px solid var(--border)',
          background: 'var(--surface)', fontFamily: 'inherit', color: 'var(--fg)',
          cursor: 'pointer', outline: 'none',
        }}>
          {services.map(s => <option key={s} value={s}>{s === 'all' ? 'all services' : s}</option>)}
        </select>

        <input type="text" placeholder="search logs..." value={searchRaw}
          onChange={e => handleSearch(e.target.value)}
          style={{
            padding: '3px 10px', fontSize: 12, border: '1px solid var(--border)',
            background: 'var(--surface)', fontFamily: 'inherit', color: 'var(--fg)',
            outline: 'none', width: 200,
          }}
          onFocus={e => (e.target.style.borderColor = 'var(--fg)')}
          onBlur={e => (e.target.style.borderColor = 'var(--border)')}
        />

        <span style={{ marginLeft: 'auto', fontSize: 11, color: 'var(--muted)' }}>
          {loading ? 'loading...' : `showing ${filtered.length} logs`}
        </span>
      </div>

      <div style={{ flex: 1, overflow: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--border)', position: 'sticky', top: 0, background: 'var(--bg)', zIndex: 5 }}>
              {['TIMESTAMP', 'LEVEL', 'SERVICE', 'MESSAGE'].map(h => (
                <th key={h} style={{ textAlign: 'left', padding: '5px 12px', fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)', fontWeight: 400, whiteSpace: 'nowrap' }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered.map((log, i) => (
              <>
                <tr key={i} onClick={() => setExpanded(expanded === i ? null : i)}
                  style={{ borderBottom: '1px solid var(--border)', background: expanded === i ? '#f5f5f5' : 'var(--surface)', cursor: 'pointer' }}
                  onMouseEnter={e => { if (expanded !== i) (e.currentTarget as HTMLTableRowElement).style.background = 'var(--surface-hover)' }}
                  onMouseLeave={e => { (e.currentTarget as HTMLTableRowElement).style.background = expanded === i ? '#f5f5f5' : 'var(--surface)' }}
                >
                  <td style={{ padding: '5px 12px', color: 'var(--muted)', whiteSpace: 'nowrap', fontSize: 11 }}>{formatTs(log.time)}</td>
                  <td style={{ padding: '5px 12px', whiteSpace: 'nowrap' }}><LevelBadge level={log.severity} /></td>
                  <td style={{ padding: '5px 12px', color: 'var(--muted)', whiteSpace: 'nowrap', fontSize: 11 }}>{log.service_name}</td>
                  <td style={{ padding: '5px 12px', maxWidth: 480, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{log.body}</td>
                </tr>
                {expanded === i && (
                  <tr key={`exp-${i}`} style={{ borderBottom: '1px solid var(--border)', background: '#f8f8f7' }}>
                    <td colSpan={4} style={{ padding: '10px 16px' }}>
                      <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '4px 16px', fontSize: 11 }}>
                        <span style={{ color: 'var(--muted)' }}>body</span><span>{log.body}</span>
                        <span style={{ color: 'var(--muted)' }}>timestamp</span><span>{log.time}</span>
                        <span style={{ color: 'var(--muted)' }}>service</span><span>{log.service_name}</span>
                        <span style={{ color: 'var(--muted)' }}>severity</span><span><LevelBadge level={log.severity} /></span>
                      </div>
                    </td>
                  </tr>
                )}
              </>
            ))}
          </tbody>
        </table>
        {!loading && filtered.length === 0 && (
          <div style={{ padding: '40px 20px', textAlign: 'center', color: 'var(--muted)', fontSize: 12 }}>
            no logs match the current filters
          </div>
        )}
      </div>
    </div>
  )
}
