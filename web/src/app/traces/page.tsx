'use client'
import { useState, useMemo, useEffect } from 'react'
import { apiFetch } from '@/lib/auth'

interface Span {
  trace_id: string
  span_id: string
  parent_id: string
  name: string
  service_name: string
  kind: string
  status_code: string
  time: string
  duration: number
}

const TIME_RANGES = [
  { label: '5M',  ms: 5 * 60 * 1000 },
  { label: '15M', ms: 15 * 60 * 1000 },
  { label: '1H',  ms: 60 * 60 * 1000 },
  { label: '6H',  ms: 6 * 60 * 60 * 1000 },
  { label: '24H', ms: 24 * 60 * 60 * 1000 },
  { label: '7D',  ms: 7 * 24 * 60 * 60 * 1000 },
]

const STATUS_LABELS: Record<string, string> = {
  STATUS_CODE_OK: 'ok',
  STATUS_CODE_ERROR: 'error',
  STATUS_CODE_UNSET: 'unset',
  '': 'unset',
}

function statusBadge(code: string) {
  const label = STATUS_LABELS[code] ?? 'unset'
  const styles: Record<string, { color: string; bg: string }> = {
    ok:    { color: '#15803d', bg: '#f0fdf4' },
    error: { color: '#b91c1c', bg: '#fef2f2' },
    unset: { color: 'rgba(0,0,0,0.35)', bg: '#f5f5f5' },
  }
  const s = styles[label] ?? styles.unset
  return <span style={{ display: 'inline-block', fontSize: 10, letterSpacing: '0.06em', color: s.color, background: s.bg, padding: '1px 6px' }}>{label}</span>
}

const SVC_COLORS = ['#1d4ed8','#15803d','#6d28d9','#b91c1c','#b45309','#0e7490','#92400e']
const colorMap: Record<string, string> = {}
let colorIdx = 0
function svcColor(s: string): string {
  if (!colorMap[s]) { colorMap[s] = SVC_COLORS[colorIdx % SVC_COLORS.length]; colorIdx++ }
  return colorMap[s]
}

function fmtDur(ns: number): string {
  const ms = ns / 1_000_000
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`
  return `${Math.round(ms)}ms`
}

function durColor(ns: number): string {
  const ms = ns / 1_000_000
  if (ms < 100) return 'var(--green)'
  if (ms < 500) return 'var(--amber)'
  return 'var(--red)'
}

function formatTs(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function Waterfall({ traceId }: { traceId: string }) {
  const [spans, setSpans] = useState<Span[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apiFetch(`/api/v1/traces/${traceId}`)
      .then(r => r.ok ? r.json() : Promise.reject(r.status))
      .then(data => setSpans(data ?? []))
      .catch(() => setSpans([]))
      .finally(() => setLoading(false))
  }, [traceId])

  if (loading) return <div style={{ padding: '12px 16px', fontSize: 11, color: 'var(--muted)' }}>loading spans...</div>
  if (!spans.length) return <div style={{ padding: '12px 16px', fontSize: 11, color: 'var(--muted)' }}>no spans found</div>

  const rootMs = Math.min(...spans.map(s => new Date(s.time).getTime()))
  const endMs = Math.max(...spans.map(s => new Date(s.time).getTime() + s.duration / 1_000_000))
  const total = endMs - rootMs || 1

  const depth: Record<string, number> = {}
  function buildDepth(pid: string, d: number) {
    spans.filter(s => s.parent_id === pid).forEach(c => { depth[c.span_id] = d; buildDepth(c.span_id, d + 1) })
  }
  const root = spans.find(s => s.parent_id === '')
  if (root) { depth[root.span_id] = 0; buildDepth(root.span_id, 1) }

  const sorted = [...spans].sort((a, b) => {
    if (a.parent_id === '') return -1
    if (b.parent_id === '') return 1
    return new Date(a.time).getTime() - new Date(b.time).getTime()
  })

  return (
    <div style={{ background: '#fafafa', borderTop: '1px solid var(--border)' }}>
      <div style={{ display: 'flex', padding: '4px 12px', borderBottom: '1px solid var(--border)', background: '#f0f0f0' }}>
        <div style={{ width: 220, fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)' }}>OPERATION</div>
        <div style={{ width: 120, fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)' }}>SERVICE</div>
        <div style={{ flex: 1, fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)' }}>TIMELINE</div>
        <div style={{ width: 70, fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)', textAlign: 'right' }}>DUR</div>
      </div>
      {sorted.map(span => {
        const color = svcColor(span.service_name)
        const offsetPct = ((new Date(span.time).getTime() - rootMs) / total) * 100
        const widthPct = ((span.duration / 1_000_000) / total) * 100
        const d = depth[span.span_id] ?? 0
        return (
          <div key={span.span_id} style={{ display: 'flex', alignItems: 'center', padding: '3px 12px', borderBottom: '1px solid var(--border)', fontSize: 11 }}>
            <div style={{ width: 220, flexShrink: 0, paddingLeft: d * 14, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', color: 'var(--muted)' }}>
              <span style={{ color, marginRight: 6 }}>▸</span>{span.name}
            </div>
            <div style={{ width: 120, flexShrink: 0, color: 'var(--muted)', fontSize: 10 }}>{span.service_name}</div>
            <div style={{ flex: 1, position: 'relative', height: 16, background: '#f0f0f0' }}>
              <div style={{ position: 'absolute', left: `${Math.min(offsetPct, 98)}%`, width: `${Math.max(widthPct, 0.5)}%`, height: '100%', background: color, opacity: 0.8 }} />
            </div>
            <div style={{ width: 70, flexShrink: 0, textAlign: 'right', color: durColor(span.duration), paddingLeft: 8 }}>{fmtDur(span.duration)}</div>
          </div>
        )
      })}
    </div>
  )
}

export default function TracesPage() {
  const [timeRangeIdx, setTimeRangeIdx] = useState(2)
  const [statusFilter, setStatusFilter] = useState('all')
  const [service, setService] = useState('all')
  const [minDuration, setMinDuration] = useState('')
  const [expanded, setExpanded] = useState<string | null>(null)
  const [traces, setTraces] = useState<Span[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const to = new Date()
    const from = new Date(to.getTime() - TIME_RANGES[timeRangeIdx].ms)
    const params = new URLSearchParams({ from: from.toISOString(), to: to.toISOString(), limit: '200' })
    setLoading(true)
    apiFetch(`/api/v1/traces?${params}`)
      .then(r => r.ok ? r.json() : Promise.reject(r.status))
      .then(data => setTraces(data ?? []))
      .catch(() => setTraces([]))
      .finally(() => setLoading(false))
  }, [timeRangeIdx])

  const services = useMemo(() => ['all', ...Array.from(new Set(traces.map(t => t.service_name))).sort()], [traces])

  const filtered = useMemo(() => {
    const minNs = minDuration ? parseInt(minDuration, 10) * 1_000_000 : 0
    return traces.filter(t => {
      if (statusFilter !== 'all' && t.status_code !== statusFilter) return false
      if (service !== 'all' && t.service_name !== service) return false
      if (minNs > 0 && t.duration < minNs) return false
      return true
    })
  }, [traces, statusFilter, service, minDuration])

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{
        position: 'sticky', top: 0, zIndex: 10, background: 'var(--bg)',
        borderBottom: '1px solid var(--border)', padding: '10px 16px',
        display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap',
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

        <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} style={{
          padding: '3px 8px', fontSize: 11, border: '1px solid var(--border)',
          background: 'var(--surface)', fontFamily: 'inherit', color: 'var(--fg)', cursor: 'pointer', outline: 'none',
        }}>
          <option value="all">all statuses</option>
          <option value="STATUS_CODE_OK">ok</option>
          <option value="STATUS_CODE_ERROR">error</option>
        </select>

        <select value={service} onChange={e => setService(e.target.value)} style={{
          padding: '3px 8px', fontSize: 11, border: '1px solid var(--border)',
          background: 'var(--surface)', fontFamily: 'inherit', color: 'var(--fg)', cursor: 'pointer', outline: 'none',
        }}>
          {services.map(s => <option key={s} value={s}>{s === 'all' ? 'all services' : s}</option>)}
        </select>

        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 11, color: 'var(--muted)' }}>min</span>
          <input type="number" placeholder="0" value={minDuration} onChange={e => setMinDuration(e.target.value)}
            style={{ width: 64, padding: '3px 8px', fontSize: 11, border: '1px solid var(--border)', background: 'var(--surface)', fontFamily: 'inherit', color: 'var(--fg)', outline: 'none' }}
            onFocus={e => (e.target.style.borderColor = 'var(--fg)')}
            onBlur={e => (e.target.style.borderColor = 'var(--border)')}
          />
          <span style={{ fontSize: 11, color: 'var(--muted)' }}>ms</span>
        </div>

        <span style={{ marginLeft: 'auto', fontSize: 11, color: 'var(--muted)' }}>
          {loading ? 'loading...' : `${filtered.length} traces`}
        </span>
      </div>

      <div style={{ flex: 1, overflow: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--border)', position: 'sticky', top: 0, background: 'var(--bg)', zIndex: 5 }}>
              {['START', 'TRACE ID', 'OPERATION', 'SERVICE', 'DURATION', 'STATUS'].map(h => (
                <th key={h} style={{ textAlign: 'left', padding: '5px 12px', fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)', fontWeight: 400, whiteSpace: 'nowrap' }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered.map(trace => (
              <>
                <tr key={trace.trace_id}
                  onClick={() => setExpanded(expanded === trace.trace_id ? null : trace.trace_id)}
                  style={{ borderBottom: expanded === trace.trace_id ? 'none' : '1px solid var(--border)', background: expanded === trace.trace_id ? '#f5f5f5' : 'var(--surface)', cursor: 'pointer' }}
                  onMouseEnter={e => { if (expanded !== trace.trace_id) (e.currentTarget as HTMLTableRowElement).style.background = 'var(--surface-hover)' }}
                  onMouseLeave={e => { (e.currentTarget as HTMLTableRowElement).style.background = expanded === trace.trace_id ? '#f5f5f5' : 'var(--surface)' }}
                >
                  <td style={{ padding: '5px 12px', color: 'var(--muted)', whiteSpace: 'nowrap', fontSize: 11 }}>{formatTs(trace.time)}</td>
                  <td style={{ padding: '5px 12px', fontFamily: 'monospace', fontSize: 11, color: 'var(--muted)', whiteSpace: 'nowrap' }}>{trace.trace_id.slice(0, 12)}...</td>
                  <td style={{ padding: '5px 12px', fontWeight: 500, whiteSpace: 'nowrap', maxWidth: 240, overflow: 'hidden', textOverflow: 'ellipsis' }}>{trace.name}</td>
                  <td style={{ padding: '5px 12px', color: svcColor(trace.service_name), whiteSpace: 'nowrap', fontSize: 11 }}>{trace.service_name}</td>
                  <td style={{ padding: '5px 12px', whiteSpace: 'nowrap', color: durColor(trace.duration), fontWeight: 500 }}>{fmtDur(trace.duration)}</td>
                  <td style={{ padding: '5px 12px' }}>{statusBadge(trace.status_code)}</td>
                </tr>
                {expanded === trace.trace_id && (
                  <tr key={`exp-${trace.trace_id}`} style={{ borderBottom: '1px solid var(--border)' }}>
                    <td colSpan={6} style={{ padding: 0 }}><Waterfall traceId={trace.trace_id} /></td>
                  </tr>
                )}
              </>
            ))}
          </tbody>
        </table>
        {!loading && filtered.length === 0 && (
          <div style={{ padding: '40px 20px', textAlign: 'center', color: 'var(--muted)', fontSize: 12 }}>
            no traces match the current filters
          </div>
        )}
      </div>
    </div>
  )
}
