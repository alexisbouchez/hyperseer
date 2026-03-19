export interface Log {
  id: string
  timestamp: string
  level: 'TRACE' | 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'FATAL'
  service_name: string
  body: string
  trace_id: string
  span_id: string
}

export interface Span {
  span_id: string
  parent_id: string | null
  operation: string
  service: string
  start_offset_ms: number
  duration_ms: number
}

export interface Trace {
  trace_id: string
  root_service: string
  root_operation: string
  start_time: string
  duration_ms: number
  status: 'ok' | 'error' | 'unset'
  span_count: number
  spans: Span[]
}

export interface DigestStats {
  error_rate: number
  req_per_min: number
  p50_ms: number
  p95_ms: number
  p99_ms: number
  top_services: ServiceStat[]
  recent_errors: Log[]
}

export interface ServiceStat {
  name: string
  req_count: number
  error_count: number
  p95_ms: number
}

const SERVICES = ['api-gateway', 'auth-service', 'payment-service', 'notification-service', 'user-service']

function randId(len = 16): string {
  return Array.from({ length: len }, () => Math.floor(Math.random() * 16).toString(16)).join('')
}

function randTraceId(): string {
  return randId(32)
}

function randSpanId(): string {
  return randId(16)
}

function randFrom<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)]
}

function tsAgo(msAgo: number): string {
  return new Date(Date.now() - msAgo).toISOString()
}

// Generate timestamps spread over last 24h
function spreadTimestamp(index: number, total: number): string {
  const span = 24 * 60 * 60 * 1000
  const ago = span - (index / total) * span + Math.random() * 60000
  return tsAgo(ago)
}

const INFO_MESSAGES: Record<string, string[]> = {
  'api-gateway': [
    'request handled in 12ms',
    'upstream routed to auth-service',
    'rate limit check passed',
    'cache hit for route /api/users',
    'request forwarded to payment-service',
    'health check passed',
    'connection pool warmed up',
    'load balancer updated weights',
  ],
  'auth-service': [
    'user authenticated successfully',
    'JWT issued for user:4821',
    'session refreshed',
    'token validated',
    'OAuth callback handled',
    'user logged out',
    'MFA check passed',
    'permission grant resolved',
  ],
  'payment-service': [
    'payment processed: $42.00',
    'invoice generated for order:9912',
    'refund issued: $15.00',
    'payment method validated',
    'transaction committed',
    'webhook dispatched to stripe',
    'idempotency key validated',
    'subscription renewed',
  ],
  'notification-service': [
    'email dispatched to user:7732',
    'push notification sent',
    'SMS delivery confirmed',
    'template rendered: order_confirmation',
    'batch notification queued: 24 recipients',
    'notification preference loaded',
    'unsubscribe processed',
    'digest scheduled',
  ],
  'user-service': [
    'user profile loaded',
    'preferences updated for user:2231',
    'avatar uploaded',
    'account created',
    'email verified',
    'password changed',
    'user data exported',
    'account deactivated',
  ],
}

const WARN_MESSAGES: Record<string, string[]> = {
  'api-gateway': [
    'slow upstream response: 820ms from auth-service',
    'rate limit approaching: 87% of quota',
    'retry attempt 2/3 for request to user-service',
    'circuit breaker half-open for payment-service',
    'connection pool near capacity: 92%',
  ],
  'auth-service': [
    'failed login attempt for user:5501 (attempt 3)',
    'token nearing expiry, refresh advised',
    'slow LDAP query: 340ms',
    'unusual login location for user:8821',
    'rate limit approaching for IP 10.0.0.42',
  ],
  'payment-service': [
    'payment gateway response time degraded: 650ms',
    'retry attempt 1/3 to stripe API',
    'high queue depth: 1200 pending transactions',
    'partial refund capped at policy limit',
    'fraud score elevated: 0.72 for txn:44812',
  ],
  'notification-service': [
    'email bounce rate elevated: 4.2%',
    'SMTP connection pool exhausted',
    'notification delayed by 45s due to backpressure',
    'retry attempt 2/3 for push notification',
    'template cache miss, falling back to DB',
  ],
  'user-service': [
    'slow query detected: 230ms on users table',
    'cache miss rate high: 42%',
    'DB connection pool at 85% capacity',
    'user export taking longer than expected',
    'stale session detected for user:9921',
  ],
}

const ERROR_MESSAGES: Record<string, string[]> = {
  'api-gateway': [
    'upstream connection refused: payment-service:8080',
    'request timeout after 5000ms',
    'circuit breaker open for auth-service',
    'TLS handshake failed with upstream',
    'invalid route configuration: missing backend',
  ],
  'auth-service': [
    'database connection failed: max retries exceeded',
    'invalid token signature',
    'JWT secret rotation failed',
    'LDAP server unreachable',
    'password hash algorithm mismatch',
  ],
  'payment-service': [
    'payment gateway timeout after 10000ms',
    'stripe API returned 500: internal error',
    'transaction rollback: deadlock detected',
    'insufficient funds: payment declined',
    'idempotency collision: duplicate request',
  ],
  'notification-service': [
    'SMTP server connection refused',
    'push notification provider returned 401',
    'template not found: order_failed',
    'message queue connection lost',
    'delivery failure after 3 retries',
  ],
  'user-service': [
    'DB write failed: unique constraint violation',
    'S3 upload failed: permission denied',
    'user not found: id=99999',
    'data export failed: timeout',
    'cache invalidation error: Redis MOVED',
  ],
}

const DEBUG_MESSAGES: Record<string, string[]> = {
  'api-gateway': ['entering request handler', 'route matched: /api/users/:id', 'headers parsed'],
  'auth-service': ['entering auth middleware', 'cache miss for token', 'fetching user from DB'],
  'payment-service': ['entering payment processor', 'building stripe payload', 'lock acquired'],
  'notification-service': ['entering notification handler', 'resolving recipient list', 'template loaded'],
  'user-service': ['entering profile handler', 'cache miss for key user:123', 'query plan selected'],
}

function getLogMessages(service: string, level: Log['level']): string[] {
  switch (level) {
    case 'INFO': return INFO_MESSAGES[service] ?? ['request handled']
    case 'WARN': return WARN_MESSAGES[service] ?? ['slow response detected']
    case 'ERROR': return ERROR_MESSAGES[service] ?? ['internal error']
    case 'DEBUG': return DEBUG_MESSAGES[service] ?? ['debug event']
    case 'TRACE': return [`span entered: handler`, `entering ${service} middleware`, `fn called: process`]
    case 'FATAL': return [`${service} crashed: out of memory`, `fatal panic: nil pointer dereference`]
  }
}

// Generate 200 logs
function generateLogs(): Log[] {
  const levels: Log['level'][] = ['TRACE', 'DEBUG', 'INFO', 'INFO', 'INFO', 'INFO', 'WARN', 'WARN', 'ERROR', 'ERROR']
  const total = 200
  return Array.from({ length: total }, (_, i) => {
    const service = randFrom(SERVICES)
    const level = randFrom(levels)
    const messages = getLogMessages(service, level)
    return {
      id: randId(12),
      timestamp: spreadTimestamp(i, total),
      level,
      service_name: service,
      body: randFrom(messages),
      trace_id: randTraceId(),
      span_id: randSpanId(),
    }
  }).sort((a, b) => b.timestamp.localeCompare(a.timestamp))
}

// Generate traces
function generateSpans(_traceId: string, rootService: string, rootOp: string): Span[] {
  const rootSpanId = randSpanId()
  const spans: Span[] = [
    {
      span_id: rootSpanId,
      parent_id: null,
      operation: rootOp,
      service: rootService,
      start_offset_ms: 0,
      duration_ms: 0, // filled later
    },
  ]

  const childCount = Math.floor(Math.random() * 4) + 2
  let maxEnd = 0

  for (let i = 0; i < childCount; i++) {
    const service = randFrom(SERVICES)
    const childOps: Record<string, string[]> = {
      'auth-service': ['verify_token', 'load_user', 'check_permissions'],
      'payment-service': ['validate_payment', 'charge_card', 'emit_receipt'],
      'notification-service': ['send_email', 'push_notification', 'queue_message'],
      'user-service': ['load_profile', 'update_preferences', 'cache_user'],
      'api-gateway': ['route_request', 'apply_rate_limit', 'log_request'],
    }
    const ops = childOps[service] ?? ['handle_request']
    const startOffset = Math.random() * 80 + i * 20
    const duration = Math.random() * 150 + 10
    maxEnd = Math.max(maxEnd, startOffset + duration)
    spans.push({
      span_id: randSpanId(),
      parent_id: rootSpanId,
      operation: randFrom(ops),
      service,
      start_offset_ms: Math.round(startOffset),
      duration_ms: Math.round(duration),
    })
  }

  const rootDuration = Math.round(maxEnd + Math.random() * 50 + 20)
  spans[0].duration_ms = rootDuration
  return spans
}

function generateTraces(): Trace[] {
  const total = 50
  const allOps = [
    'POST /api/orders',
    'GET /api/users/:id',
    'PUT /api/payments',
    'PATCH /api/profile',
    'GET /api/products',
    'DELETE /api/sessions',
    'POST /api/checkout',
    'GET /api/dashboard',
    'POST /api/auth/login',
    'GET /api/notifications',
  ]
  return Array.from({ length: total }, (_, i) => {
    const rootService = 'api-gateway'
    const rootOp = randFrom(allOps)
    const isError = Math.random() < 0.15
    const spans = generateSpans(randTraceId(), rootService, rootOp)
    const duration = spans[0].duration_ms
    return {
      trace_id: randTraceId(),
      root_service: rootService,
      root_operation: rootOp,
      start_time: spreadTimestamp(i, total),
      duration_ms: duration,
      status: (isError ? 'error' : 'ok') as Trace['status'],
      span_count: spans.length,
      spans,
    }
  }).sort((a, b) => b.start_time.localeCompare(a.start_time))
}

export const MOCK_LOGS: Log[] = generateLogs()
export const MOCK_TRACES: Trace[] = generateTraces()

function buildDigest(): DigestStats {
  const errors = MOCK_LOGS.filter(l => l.level === 'ERROR' || l.level === 'FATAL')
  const errorRate = Math.round((errors.length / MOCK_LOGS.length) * 100 * 10) / 10

  const durations = MOCK_TRACES.map(t => t.duration_ms).sort((a, b) => a - b)
  const p50 = durations[Math.floor(durations.length * 0.5)]
  const p95 = durations[Math.floor(durations.length * 0.95)]
  const p99 = durations[Math.floor(durations.length * 0.99)]

  const serviceStats: Record<string, { reqs: number; errors: number; durations: number[] }> = {}
  for (const s of SERVICES) {
    serviceStats[s] = { reqs: 0, errors: 0, durations: [] }
  }
  for (const log of MOCK_LOGS) {
    if (serviceStats[log.service_name]) {
      serviceStats[log.service_name].reqs++
      if (log.level === 'ERROR' || log.level === 'FATAL') serviceStats[log.service_name].errors++
    }
  }
  for (const trace of MOCK_TRACES) {
    for (const span of trace.spans) {
      if (serviceStats[span.service]) {
        serviceStats[span.service].durations.push(span.duration_ms)
      }
    }
  }

  const top_services: ServiceStat[] = SERVICES.map(name => {
    const s = serviceStats[name]
    const sorted = s.durations.sort((a, b) => a - b)
    const p95idx = Math.floor(sorted.length * 0.95)
    return {
      name,
      req_count: s.reqs,
      error_count: s.errors,
      p95_ms: sorted[p95idx] ?? 0,
    }
  }).sort((a, b) => b.req_count - a.req_count)

  return {
    error_rate: errorRate,
    req_per_min: Math.round(MOCK_LOGS.length / (24 * 60)),
    p50_ms: p50 ?? 0,
    p95_ms: p95 ?? 0,
    p99_ms: p99 ?? 0,
    top_services,
    recent_errors: errors.slice(0, 10),
  }
}

export const MOCK_DIGEST: DigestStats = buildDigest()
