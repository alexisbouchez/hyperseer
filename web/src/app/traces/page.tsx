'use client'
import { useState, useMemo } from 'react'
import { MOCK_TRACES } from '@/lib/mock'
import type { Trace, Span } from '@/lib/mock'

const TIME_RANGES = [
  { label: '5M',  ms: 5 * 60 * 1000 },
  { label: '15M', ms: 15 * 60 * 1000 },
  { label: '1H',  ms: 60 * 60 * 1000 },
  { label: '6H',  ms: 6 * 60 * 60 * 1000 },
  { label: '24H', ms: 24 * 60 * 60 * 1000 },
  { label: '7D',  ms: 7 * 24 * 60 * 60 * 1000 },
]

const SERVICES = ['all', 'api-gateway', 'auth-service', 'payment-service', 'notification-service', 'user-service']
const STATUSES = ['all', 'ok', 'error', 'unset']

const SERVICE_COLORS: Record<string, string> = {
  'api-gateway':        '#1d4ed8',
  'auth-service':       '#15803d',
  'payment-service':    '#6d28d9',
  'notification-service': '#b91c1c',
  'user-service':       '#b45309',
}

function getServiceColor(service: string): string {
  return SERVICE_COLORS[service] ?? 'rgba(0,0,0,0.35)'
}

function fmtDuration(ms: number): string {
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`
  return `${ms}ms`
}

function durationColor(ms: number): string {
  if (ms < 100) return 'var(--green)'
  if (ms < 500) return 'var(--amber)'
  return 'var(--red)'
}

function statusBadge(status: Trace['status']) {
  const styles: Record<Trace['status'], { color: string; bg: string }> = {
    ok:    { color: '#15803d', bg: '#f0fdf4' },
    error: { color: '#b91c1c', bg: '#fef2f2' },
    unset: { color: 'rgba(0,0,0,0.35)', bg: '#f5f5f5' },
  }
  const s = styles[status]
  return (
    <span style={{
      display: 'inline-block',
      fontSize: 10, letterSpacing: '0.06em',
      color: s.color, background: s.bg,
      padding: '1px 6px',
    }}>
      {status}
    </span>
  )
}

function formatTs(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

// Waterfall span row
function SpanRow({ span, totalDuration, depth }: { span: Span; totalDuration: number; depth: number }) {
  const color = getServiceColor(span.service)
  const widthPct = totalDuration > 0 ? (span.duration_ms / totalDuration) * 100 : 1
  const offsetPct = totalDuration > 0 ? (span.start_offset_ms / totalDuration) * 100 : 0

  return (
    <div style={{ display: 'flex', alignItems: 'center', padding: '3px 12px', borderBottom: '1px solid var(--border)', fontSize: 11 }}>
      <div style={{ width: 200, flexShrink: 0, paddingLeft: depth * 16, color: 'var(--muted)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
        <span style={{ color, marginRight: 6 }}>▸</span>
        {span.operation}
      </div>
      <div style={{ width: 120, flexShrink: 0, color: 'var(--muted)', fontSize: 10 }}>{span.service}</div>
      <div style={{ flex: 1, position: 'relative', height: 16, background: '#f0f0f0' }}>
        <div style={{
          position: 'absolute',
          left: `${Math.min(offsetPct, 98)}%`,
          width: `${Math.max(widthPct, 0.5)}%`,
          height: '100%',
          background: color,
          opacity: 0.8,
        }} />
      </div>
      <div style={{ width: 70, flexShrink: 0, textAlign: 'right', color: durationColor(span.duration_ms), paddingLeft: 8, fontSize: 11 }}>
        {fmtDuration(span.duration_ms)}
      </div>
    </div>
  )
}

// Expanded waterfall for a trace
function Waterfall({ trace }: { trace: Trace }) {
  // Build depth map: root = 0, children of root = 1, etc.
  const depthMap: Record<string, number> = {}
  function buildDepth(spanId: string | null, depth: number) {
    const children = trace.spans.filter(s => s.parent_id === spanId)
    for (const child of children) {
      depthMap[child.span_id] = depth
      buildDepth(child.span_id, depth + 1)
    }
  }
  const root = trace.spans.find(s => s.parent_id === null)
  if (root) {
    depthMap[root.span_id] = 0
    buildDepth(root.span_id, 1)
  }

  // Sort: root first, then by start_offset_ms
  const sorted = [...trace.spans].sort((a, b) => {
    if (a.parent_id === null) return -1
    if (b.parent_id === null) return 1
    return a.start_offset_ms - b.start_offset_ms
  })

  return (
    <div style={{ background: '#fafafa', borderTop: '1px solid var(--border)' }}>
      <div style={{ display: 'flex', alignItems: 'center', padding: '4px 12px', borderBottom: '1px solid var(--border)', background: '#f0f0f0' }}>
        <div style={{ width: 200, fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)' }}>OPERATION</div>
        <div style={{ width: 120, fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)' }}>SERVICE</div>
        <div style={{ flex: 1, fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)' }}>TIMELINE</div>
        <div style={{ width: 70, fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)', textAlign: 'right' }}>DUR</div>
      </div>
      {sorted.map(span => (
        <SpanRow
          key={span.span_id}
          span={span}
          totalDuration={trace.duration_ms}
          depth={depthMap[span.span_id] ?? 0}
        />
      ))}
    </div>
  )
}

export default function TracesPage() {
  const [timeRangeIdx, setTimeRangeIdx] = useState(4)
  const [status, setStatus] = useState('all')
  const [service, setService] = useState('all')
  const [minDuration, setMinDuration] = useState('')
  const [expanded, setExpanded] = useState<string | null>(null)

  const filtered = useMemo(() => {
    const cutoff = Date.now() - TIME_RANGES[timeRangeIdx].ms
    const minMs = minDuration ? parseInt(minDuration, 10) : 0
    return MOCK_TRACES.filter(t => {
      if (new Date(t.start_time).getTime() < cutoff) return false
      if (status !== 'all' && t.status !== status) return false
      if (service !== 'all' && t.root_service !== service) return false
      if (minMs > 0 && t.duration_ms < minMs) return false
      return true
    })
  }, [timeRangeIdx, status, service, minDuration])

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {/* filter bar */}
      <div style={{
        position: 'sticky', top: 0, zIndex: 10,
        background: 'var(--bg)', borderBottom: '1px solid var(--border)',
        padding: '10px 16px',
        display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap',
      }}>
        {/* time range */}
        <div style={{ display: 'flex', gap: 0 }}>
          {TIME_RANGES.map((r, i) => (
            <button
              key={r.label}
              onClick={() => setTimeRangeIdx(i)}
              style={{
                padding: '3px 8px', fontSize: 11, letterSpacing: '0.06em',
                border: '1px solid var(--border)',
                marginLeft: i === 0 ? 0 : -1,
                background: i === timeRangeIdx ? 'var(--fg)' : 'var(--surface)',
                color: i === timeRangeIdx ? '#fff' : 'var(--muted)',
                cursor: 'pointer', fontFamily: 'inherit',
              }}
            >
              {r.label}
            </button>
          ))}
        </div>

        {/* status */}
        <select
          value={status}
          onChange={e => setStatus(e.target.value)}
          style={{
            padding: '3px 8px', fontSize: 11, border: '1px solid var(--border)',
            background: 'var(--surface)', fontFamily: 'inherit', color: 'var(--fg)',
            cursor: 'pointer', outline: 'none',
          }}
        >
          {STATUSES.map(s => <option key={s} value={s}>{s === 'all' ? 'all statuses' : s}</option>)}
        </select>

        {/* service */}
        <select
          value={service}
          onChange={e => setService(e.target.value)}
          style={{
            padding: '3px 8px', fontSize: 11, border: '1px solid var(--border)',
            background: 'var(--surface)', fontFamily: 'inherit', color: 'var(--fg)',
            cursor: 'pointer', outline: 'none',
          }}
        >
          {SERVICES.map(s => <option key={s} value={s}>{s === 'all' ? 'all services' : s}</option>)}
        </select>

        {/* min duration */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 11, color: 'var(--muted)' }}>min</span>
          <input
            type="number"
            placeholder="0"
            value={minDuration}
            onChange={e => setMinDuration(e.target.value)}
            style={{
              width: 64, padding: '3px 8px', fontSize: 11, border: '1px solid var(--border)',
              background: 'var(--surface)', fontFamily: 'inherit', color: 'var(--fg)', outline: 'none',
            }}
            onFocus={e => (e.target.style.borderColor = 'var(--fg)')}
            onBlur={e => (e.target.style.borderColor = 'var(--border)')}
          />
          <span style={{ fontSize: 11, color: 'var(--muted)' }}>ms</span>
        </div>

        <span style={{ marginLeft: 'auto', fontSize: 11, color: 'var(--muted)' }}>
          {filtered.length} traces
        </span>
      </div>

      {/* trace table */}
      <div style={{ flex: 1, overflow: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--border)', position: 'sticky', top: 0, background: 'var(--bg)', zIndex: 5 }}>
              {['START', 'TRACE ID', 'ROOT OPERATION', 'SERVICE', 'DURATION', 'SPANS', 'STATUS'].map(h => (
                <th key={h} style={{
                  textAlign: 'left', padding: '5px 12px',
                  fontSize: 10, letterSpacing: '0.1em', color: 'var(--muted)', fontWeight: 400,
                  whiteSpace: 'nowrap',
                }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered.map((trace) => (
              <>
                <tr
                  key={trace.trace_id}
                  onClick={() => setExpanded(expanded === trace.trace_id ? null : trace.trace_id)}
                  style={{
                    borderBottom: expanded === trace.trace_id ? 'none' : '1px solid var(--border)',
                    background: expanded === trace.trace_id ? '#f5f5f5' : 'var(--surface)',
                    cursor: 'pointer',
                  }}
                  onMouseEnter={e => { if (expanded !== trace.trace_id) (e.currentTarget as HTMLTableRowElement).style.background = 'var(--surface-hover)' }}
                  onMouseLeave={e => { (e.currentTarget as HTMLTableRowElement).style.background = expanded === trace.trace_id ? '#f5f5f5' : 'var(--surface)' }}
                >
                  <td style={{ padding: '5px 12px', color: 'var(--muted)', whiteSpace: 'nowrap', fontSize: 11 }}>
                    {formatTs(trace.start_time)}
                  </td>
                  <td style={{ padding: '5px 12px', fontFamily: 'monospace', fontSize: 11, color: 'var(--muted)', whiteSpace: 'nowrap' }}>
                    {trace.trace_id.slice(0, 12)}…
                  </td>
                  <td style={{ padding: '5px 12px', fontWeight: 500, whiteSpace: 'nowrap', maxWidth: 240, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {trace.root_operation}
                  </td>
                  <td style={{ padding: '5px 12px', color: getServiceColor(trace.root_service), whiteSpace: 'nowrap', fontSize: 11 }}>
                    {trace.root_service}
                  </td>
                  <td style={{ padding: '5px 12px', whiteSpace: 'nowrap', color: durationColor(trace.duration_ms), fontWeight: 500 }}>
                    {fmtDuration(trace.duration_ms)}
                  </td>
                  <td style={{ padding: '5px 12px', color: 'var(--muted)' }}>
                    {trace.span_count}
                  </td>
                  <td style={{ padding: '5px 12px' }}>
                    {statusBadge(trace.status)}
                  </td>
                </tr>
                {expanded === trace.trace_id && (
                  <tr key={`${trace.trace_id}-exp`} style={{ borderBottom: '1px solid var(--border)' }}>
                    <td colSpan={7} style={{ padding: 0 }}>
                      <Waterfall trace={trace} />
                    </td>
                  </tr>
                )}
              </>
            ))}
          </tbody>
        </table>
        {filtered.length === 0 && (
          <div style={{ padding: '40px 20px', textAlign: 'center', color: 'var(--muted)', fontSize: 12 }}>
            no traces match the current filters
          </div>
        )}
      </div>
    </div>
  )
}
